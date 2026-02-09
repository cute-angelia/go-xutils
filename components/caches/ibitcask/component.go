package ibitcask

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"git.mills.io/prologic/bitcask"
)

var ErrClosed = errors.New("datastore closed")

var iComponent *Component

const PackageName = "component.ibitcask"

type Component struct {
	config *config
	db     *bitcask.Bitcask

	// close
	closed    bool
	closeLk   sync.RWMutex
	closeOnce sync.Once
	closing   chan struct{}
}

func GetComponent() *Component {
	if iComponent == nil {
		panic("ibitcask is need init")
	}
	return iComponent
}

// newComponent ...
func newComponent(config *config) *Component {
	comp := &Component{
		config:  config,
		closing: make(chan struct{}),
	}

	// db
	if db, err := bitcask.Open(config.Path, bitcask.WithMaxDatafileSize(config.MaxDatafileSize)); err != nil {
		panic(fmt.Sprintf("[%s] 初始化失败: %s", PackageName, err))
	} else {
		comp.db = db
	}

	if config.GcInterval > 0 {
		go comp.periodicGC()
	}

	// 赋值
	iComponent = comp

	return comp
}

// Keep scheduling GC's AFTER `gcInterval` has passed since the previous GC
func (d *Component) periodicGC() {
	gcTimeout := time.NewTimer(d.config.GcInterval)
	defer gcTimeout.Stop()

	for {
		select {
		case <-gcTimeout.C:
			if err := d.db.RunGC(); err != nil {
				log.Println(err)
			}
			gcTimeout.Reset(d.config.GcInterval)
		case <-d.closing:
			return
		}
	}
}

// ======= functions =======

func (d *Component) GetDb() (*bitcask.Bitcask, error) {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed { // Use d, not iComponent
		return nil, ErrClosed
	}
	return d.db, nil
}

// Close 关闭db
func (d *Component) Close() error {
	d.closeLk.Lock() // 加写锁
	defer d.closeLk.Unlock()

	if d.closed {
		return ErrClosed
	}

	d.closeOnce.Do(func() {
		close(d.closing)
	})

	d.closed = true
	return d.db.Close()
}

func (c *Component) GenerateCacheKey(bucket string, key string) string {
	return fmt.Sprintf("%s:%s", bucket, key)
}

// Set 保存数据
func (d *Component) Set(key string, value string, ttl time.Duration) error {
	if d.closed {
		return ErrClosed
	}
	return d.db.PutWithTTL([]byte(key), []byte(value), ttl)
}

// SetWithBucket 保存数据
func (d *Component) SetWithBucket(bucket string, key string, value string, ttl time.Duration) error {
	if d.closed {
		return ErrClosed
	}
	// 修复：使用生成的组合键
	realKey := d.GenerateCacheKey(bucket, key)
	return d.db.PutWithTTL([]byte(realKey), []byte(value), ttl)
}

// Get 获取数据
func (d *Component) Get(key string) (string, error) {
	if d.closed {
		return "", ErrClosed
	}

	if key == "" {
		return "", errors.New("empty key")
	}

	v, err := d.db.Get([]byte(key))
	return string(v), err
}

func (d *Component) Delete(key string) error {
	if d.closed {
		return ErrClosed
	}
	return d.db.Delete([]byte(key))
}

func (d *Component) GetMulti(keys []string) map[string]string {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()

	result := make(map[string]string)
	if d.closed {
		return result
	}

	for _, key := range keys {
		if value, err := d.db.Get([]byte(key)); err == nil {
			result[key] = string(value)
		}
	}
	return result
}

func (d *Component) Contains(key string) bool {
	return d.db.Has([]byte(key))
}

func (d *Component) Flush() error {
	return d.db.DeleteAll()
}

func (d *Component) Scan(prefix string, f func(key string) error) (err error) {
	return d.db.Scan([]byte(prefix), func(key []byte) error {
		return f(string(key))
	})
}

func (d *Component) Fold(f func(key string) error) (err error) {
	return d.db.Fold(func(key []byte) error {
		return f(string(key))
	})
}
