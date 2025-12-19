package ibunt

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cute-angelia/go-xutils/syntax/ijson"
	"github.com/tidwall/buntdb"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var BuntCaches sync.Map
var iComponent *Component

const PackageName = "component.ibunt"

type Component struct {
	config *config
}

func GetComponent(dbname string) *Component {
	iComponent.config.Name = dbname
	return iComponent
}

// newComponent ...
func newComponent(config *config) *Component {
	comp := &Component{}
	comp.config = config

	//  自动初始化， 有些人经常忘。。。
	if err := comp.initBuntDb(); err != nil {
		log.Println(fmt.Sprintf("[%s] 初始化失败", PackageName))
	}

	// 赋值
	iComponent = comp

	return comp
}

// initBuntDb 初始化
func (c *Component) initBuntDb() error {
	if _, ok := BuntCaches.Load(c.config.Name); ok {
		return nil
	} else {
		// dbname
		db := c.config.DbFile
		if !strings.Contains(db, ".db") {
			db = db + ".db"
		}
		// 无法创建
		if cache, err := buntdb.Open(db); err != nil {
			// 移除异常 db
			if _, e := os.Stat(db); e == nil {
				os.Remove(db)

				// retry 重新打开一次
				if cache, err := buntdb.Open(db); err != nil {
					panic(err)
				} else {
					cache.SetConfig(buntdb.Config{
						AutoShrinkDisabled:   true,
						AutoShrinkMinSize:    40,
						AutoShrinkPercentage: 30,
					})
					cache.Shrink()
					SetDb(c.config.Name, cache)
				}
			}
		} else {
			cache.SetConfig(buntdb.Config{
				AutoShrinkDisabled:   true,
				AutoShrinkMinSize:    40,
				AutoShrinkPercentage: 30,
			})
			cache.Shrink()
			SetDb(c.config.Name, cache)
		}
		// 初始化日志
		log.Println(fmt.Sprintf("[%s] Name:%s, dbFile=%s 初始化",
			PackageName,
			c.config.Name,
			c.config.DbFile,
		))
		return nil
	}
}

func SetConfig(db *buntdb.DB, conf buntdb.Config) {
	if conf.AutoShrinkMinSize > 0 {
		conf.AutoShrinkDisabled = true
		err := db.SetConfig(conf)
		if err != nil {
			log.Println("buntdb.DB SetConfig error", err)
		}
	} else {
		err := db.SetConfig(buntdb.Config{
			AutoShrinkDisabled:   true,
			AutoShrinkMinSize:    50,
			AutoShrinkPercentage: 100,
			SyncPolicy:           buntdb.Always,
		})
		if err != nil {
			log.Println("buntdb.DB SetConfig error", err)
		} else {
			var config buntdb.Config
			if err := db.ReadConfig(&config); err != nil {
				log.Println("buntdb.DB GetConfig error", err)
			} else {
				// log.Println("ReadConfig", config)
			}
		}
	}
}

func GetDb(name string) *buntdb.DB {
	if v, ok := BuntCaches.Load(name); ok {
		return v.(*buntdb.DB)
	} else {
		log.Println("buntdb.未初始化:" + name)
		return nil
	}
}

func SetDb(name string, db *buntdb.DB) {
	SetConfig(db, buntdb.Config{})
	BuntCaches.Store(name, db)
}

func Set(dbname string, key string, val string, ttl time.Duration) error {
	if db := GetDb(dbname); db != nil {
		db.Update(func(tx *buntdb.Tx) error {
			tx.Set(key, val, &buntdb.SetOptions{Expires: true, TTL: ttl})
			return nil
		})
		return nil
	} else {
		return fmt.Errorf("无法找到 db")
	}
}

// GetOrSet 获取数据或者存储数据
func GetOrSet(dbname string, key string, function func() interface{}, ttl time.Duration) (string, error) {
	cacheData := Get(dbname, key)
	if len(cacheData) > 10 {
		return cacheData, nil
	} else {
		byteJson, _ := ijson.Marshal(function())
		strJson := string(byteJson)
		err := Set(dbname, key, strJson, ttl)
		return strJson, err
	}
}

func Get(dbname string, key string) string {
	if db := GetDb(dbname); db != nil {
		val := ""
		db.View(func(tx *buntdb.Tx) error {
			val, _ = tx.Get(key)
			return nil
		})

		if len(val) == 0 {
			return ""
		} else {
			return val
		}
	} else {
		log.Println(fmt.Errorf("无法找到 db" + dbname))
		return ""
	}
}

// SetAny 保存结构型数据
func SetAny(dbname string, key string, val any, ttl time.Duration) error {
	if vals, err := json.Marshal(val); err != nil {
		return errors.New("SetAny:" + key + ":" + err.Error())
	} else {
		return Set(dbname, key, string(vals), ttl)
	}
}

// GetAny 获取数据，并赋值给v
func GetAny(dbname string, key string, v any) error {
	result := Get(dbname, key)
	if len(result) > 0 {
		return json.Unmarshal([]byte(result), v)
	} else {
		return errors.New("GetAny:" + key + ":缓存内容不存在")
	}
}

func ShowTttl(dbname string, key string) string {
	if db := GetDb(dbname); db != nil {
		val := ""
		db.View(func(tx *buntdb.Tx) error {
			val, _ = tx.Get(key)

			itemTTL, err := tx.TTL(key)
			if err == nil {
				exat := time.Now().Add(itemTTL)
				val = exat.Format(time.RFC3339)
			}
			return nil
		})

		if len(val) == 0 {
			return ""
		} else {
			return val
		}
	} else {
		log.Println(fmt.Errorf("无法找到 db" + dbname))
		return ""
	}
}

func Delete(dbname string, key string) error {
	if db := GetDb(dbname); db != nil {
		return db.Update(func(tx *buntdb.Tx) error {
			_, err := tx.Delete(key)
			return err
		})
	} else {
		log.Println(fmt.Errorf("无法找到 db" + dbname))
		return fmt.Errorf("无法找到 db" + dbname)
	}
}

func (c *Component) GenerateCacheKey(bucket string, key string) string {
	return fmt.Sprintf("%s:%s", bucket, key)
}

func (c *Component) Get(key string) (string, error) {
	return Get(c.config.Name, key), nil
}

func (c *Component) GetMulti(keys []string) map[string]string {
	result := make(map[string]string)
	for _, key := range keys {
		result[key] = Get(c.config.Name, key)
	}
	return result
}

func (c *Component) Set(key string, value string, ttl time.Duration) error {
	return Set(c.config.Name, key, value, ttl)
}

func (c *Component) SetWithBucket(bucket string, key string, value string, ttl time.Duration) error {
	return Set(c.config.Name, key, value, ttl)
}

func (c *Component) Contains(key string) bool {
	v, _ := c.Get(key)
	return len(v) > 0
}

func (c *Component) Delete(key string) error {
	return Delete(c.config.Name, key)
}

func (c *Component) Flush() error {
	return GetDb(c.config.Name).Shrink()
}

func (c *Component) Scan(prefix string, f func(key string) error) (err error) {
	if db := GetDb(c.config.Name); db != nil {
		return db.View(func(tx *buntdb.Tx) error {
			err := tx.Ascend(prefix, func(key, value string) bool {
				f(key)
				return true
			})
			return err
		})
	} else {
		return errors.New(fmt.Sprintf("无法找到 db" + c.config.Name))
	}
}

func (c *Component) Fold(f func(key string) error) (err error) {
	panic("implement me")
}
