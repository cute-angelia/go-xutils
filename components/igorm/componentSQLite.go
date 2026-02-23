package igorm

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const PackageNameSQLite = "component.igorm.sqlite"

// GetGormSQLite 获取 gorm.DB 对象
func GetGormSQLite(dbName string) (*gorm.DB, error) {
	// 统一从 gormPoolSQLite 读取
	if v, ok := gormPoolSQLite.Load(dbName); ok {
		return v.(*gorm.DB), nil
	}
	return nil, errors.New(PackageNameSQLite + " 获取失败:" + dbName + " 未初始化")
}

// MustInitSqlite 初始化
func (c *Component) MustInitSqlite() *Component {
	// 1. 检查是否已初始化，避免重复 Open
	if _, ok := gormPoolSQLite.Load(c.config.DbName); ok {
		return c
	}

	// 2. 确定数据库路径
	pathdb := c.config.DbFile
	if len(pathdb) == 0 {
		pathdb = fmt.Sprintf("./%s_SQLite.db", c.config.DbName) // 建议用 ./ 相对路径
	}

	if c.config.Debug {
		log.Println(PackageNameSQLite, "sqlite path:", pathdb)
	}

	// 3. 执行初始化
	db := c.initSqliteDb(pathdb)
	if db == nil {
		panic(fmt.Sprintf("[%s] ❌数据库 [%s] 初始化失败", PackageNameSQLite, c.config.DbName))
	}

	gormPoolSQLite.Store(c.config.DbName, db)

	log.Printf("[%s] Name:%s 初始化成功 (Path:%s)", PackageNameSQLite, c.config.DbName, pathdb)
	return c
}

func (c *Component) initSqliteDb(pathdb string) *gorm.DB {
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

	// 1. 尝试打开 (SQLite 实际上很少有网络超时，主要是 IO 错误)
	dsn := fmt.Sprintf("%s?_busy_timeout=5000", pathdb)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		log.Printf("[%s] ❌数据库连接异常 Path:%s, Err:%v", PackageNameSQLite, pathdb, err)
		return nil
	}

	// 【修改 2】开启 WAL 模式 (非常重要)
	// 这一步必须在 Open 之后立即执行
	db.Exec("PRAGMA journal_mode=WAL;")

	// 2. 配置底层连接池
	idb, err := db.DB()
	if err != nil {
		return nil
	}

	// 【重要】针对 SQLite 的特殊优化：
	// 如果是写频繁的应用，建议 MaxOpenConns 设为 1，防止 database is locked
	if c.config.MaxOpenConns > 0 {
		idb.SetMaxOpenConns(c.config.MaxOpenConns)
	} else {
		// 写频繁建议设为 1，但在 WAL 模式下可以适当放宽到 2-5
		idb.SetMaxOpenConns(3)
	}

	idb.SetMaxIdleConns(c.config.MaxIdleConns)
	idb.SetConnMaxLifetime(c.config.MaxLifetime)

	return db
}
