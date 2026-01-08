package igorm

import (
	"errors"
	"fmt"
	"gorm.io/driver/postgres" // 需安装: go get gorm.io/driver/postgres
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"regexp"
	"time"
)

/**
 * @Description: PostgreSQL 组件
 * @Author: CuteAngelia
 * @Date: 2024-09-10 15:40:40
 * @LastEditors: CuteAngelia
 * @LastEditTime: 2024-09-10 15:40:40

PostgreSQL 使用的是 Key=Value 格式。
MySQL: user:pass@tcp(host:port)/dbname?timeout=5s
PostgreSQL: host=127.0.0.1 user=root password=123 dbname=test port=5432 sslmode=disable connect_timeout=5
*/

const PackageNamePostgres = "component.igorm.postgres"

// GetGormPostgres 获取 PostgreSQL gorm.DB 对象
func GetGormPostgres(dbName string) (*gorm.DB, error) {
	if v, ok := gormPoolPostgreSQL.Load(dbName); ok {
		return v.(*gorm.DB), nil
	} else {
		return nil, errors.New(PackageNamePostgres + " 获取失败:" + dbName + " 未初始化")
	}
}

// MustInitPostgres 初始化 PostgreSQL
func (c *Component) MustInitPostgres() *Component {
	// 1. 配置校验
	if len(c.config.Dsn) == 0 || len(c.config.DbName) == 0 {
		panic(fmt.Sprintf("❌数据库配置不正确 dbName=%s dsn=%s", c.config.DbName, c.config.Dsn))
	}

	// 2. 初始化 db
	if _, ok := gormPoolPostgreSQL.Load(c.config.DbName); !ok {
		db := c.initPostgresDb()
		if db == nil {
			panic(fmt.Sprintf("❌PostgreSQL [%s] 初始化失败", c.config.DbName))
		}
		gormPoolPostgreSQL.Store(c.config.DbName, db)
	}

	log.Println(fmt.Sprintf("[%s] Name:%s 初始化成功",
		PackageNamePostgres,
		c.config.DbName,
	))

	return c
}

func (c *Component) initPostgresDb() *gorm.DB {
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
	// 注意：PostgreSQL 的 DSN 示例: "host=localhost user=gorm password=gorm dbname=gorm port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(c.config.Dsn), &gorm.Config{
		Logger: newLogger,
	})

	// 2. 错误拦截
	if err != nil {
		// 针对 PostgreSQL DSN 的正则匹配提取 Host（通常匹配 host=xxx 或地址部分）
		re := regexp.MustCompile(`host=([^\s]+)`)
		match := re.FindStringSubmatch(c.config.Dsn)
		host := "unknown"
		if len(match) > 1 {
			host = match[1]
		}
		log.Printf("[%s] ❌数据库连接异常 Host:%s, DB:%s, Err:%v", PackageNamePostgres, host, c.config.DbName, err)
		return nil
	}

	// 3. 获取底层池并配置
	idb, err := db.DB()
	if err != nil {
		log.Printf("[%s] ❌获取底层DB对象失败: %v", PackageNamePostgres, err)
		return nil
	}

	// 4. 设置连接池参数
	idb.SetMaxIdleConns(c.config.MaxIdleConns)
	idb.SetMaxOpenConns(c.config.MaxOpenConns)
	idb.SetConnMaxLifetime(c.config.MaxLifetime)

	return db
}
