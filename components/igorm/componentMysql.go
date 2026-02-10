package igorm

import (
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"regexp"
	"time"
)

const PackageNameMysql = "component.igorm.mysql"

// GetGormMysql 获取 gorm.DB 对象
func GetGormMysql(dbName string) (*gorm.DB, error) {
	if v, ok := gormPool.Load(dbName); ok {
		return v.(*gorm.DB), nil
	} else {
		return nil, errors.New(packageName + " 获取失败:" + dbName + " 未初始化")
	}
}

// MustInitMysql 初始化
func (c *Component) MustInitMysql() *Component {
	// 配置必须信息
	if len(c.config.Dsn) == 0 || len(c.config.DbName) == 0 {
		panic(fmt.Sprintf("❌数据库配置不正确 dbName=%s dsn=%s", c.config.DbName, c.config.Dsn))
	}
	// 初始化 db
	if _, ok := gormPool.Load(c.config.DbName); !ok {
		db := c.initMysqlDb()
		if db == nil {
			panic(fmt.Sprintf("❌数据库 [%s] 初始化失败", c.config.DbName))
		}
		gormPool.Store(c.config.DbName, db)
	}

	// 初始化日志
	log.Println(fmt.Sprintf("[%s] Name:%s 初始化",
		PackageNameMysql,
		c.config.DbName,
	))

	return c
}

func (c *Component) InitMysql() *Component {
	// 配置必须信息
	if len(c.config.Dsn) == 0 || len(c.config.DbName) == 0 {
		panic(fmt.Sprintf("❌数据库配置不正确 dbName=%s dsn=%s", c.config.DbName, c.config.Dsn))
	}
	// 初始化 db
	if _, ok := gormPool.Load(c.config.DbName); !ok {
		db := c.initMysqlDb()
		if db == nil {
			log.Println(fmt.Sprintf("❌数据库 [%s] 初始化失败", c.config.DbName))
		} else {
			gormPool.Store(c.config.DbName, db)
		}
	}

	// 初始化日志
	log.Println(fmt.Sprintf("[%s] Name:%s 初始化",
		PackageNameMysql,
		c.config.DbName,
	))

	return c
}

func (c *Component) initMysqlDb() *gorm.DB {
	var vlog *log.Logger
	if c.config.LoggerWriter == nil {
		vlog = log.New(os.Stdout, "\r\n", log.LstdFlags|log.Lshortfile)
	} else {
		vlog = log.New(c.config.LoggerWriter, "", 0)
	}

	newLogger := logger.New(
		vlog,
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      c.config.LogLevel,
			Colorful:      true,
		},
	)

	// 1. 尝试打开连接
	db, err := gorm.Open(mysql.Open(c.config.Dsn), &gorm.Config{
		Logger: newLogger,
	})

	// 2. 核心改进：先判断 err，如果连接失败立刻返回 nil
	if err != nil {
		re := regexp.MustCompile(`tcp\((.*?)\)`)
		match := re.FindStringSubmatch(c.config.Dsn)
		host := "unknown"
		if len(match) > 1 {
			host = match[1]
		}
		log.Printf("[%s] ❌数据库连接异常 Host:%s, DB:%s, Err:%v", packageName, host, c.config.DbName, err)
		return nil // 发现错误立即返回，避免后续 db.DB() 崩溃
	}

	// 3. 获取底层连接池进行配置
	idb, err := db.DB()
	if err != nil {
		log.Printf("[%s] ❌获取底层DB对象失败: %v", packageName, err)
		return nil
	}

	// 4. 设置连接池参数
	idb.SetMaxIdleConns(c.config.MaxIdleConns)   // ==>  用于设置连接池中空闲连接的最大数量(10)
	idb.SetMaxOpenConns(c.config.MaxOpenConns)   // ==>  设置打开数据库连接的最大数量(100)
	idb.SetConnMaxLifetime(c.config.MaxLifetime) // 最大时间

	return db
}
