package igorm

import (
	"errors"
	"fmt"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"log"
)

const PackageNameSqlServer = "component.igorm.sqlserver"

func GetGormSqlServer(dbName string) (*gorm.DB, error) {
	if v, ok := gormPool.Load(dbName); ok {
		return v.(*gorm.DB), nil
	}
	return nil, errors.New(PackageNameSqlServer + " 获取失败: " + dbName + " 未初始化")
}

func (c *Component) MustInitSqlServer() *Component {
	if len(c.config.Dsn) == 0 {
		panic("❌ SQL Server DSN 缺失")
	}
	if _, ok := gormPool.Load(c.config.DbName); !ok {
		db := c.initSqlServerDb()
		if db == nil {
			panic(fmt.Sprintf("❌ SQL Server [%s] 启动失败", c.config.DbName))
		}
		gormPool.Store(c.config.DbName, db)
	}
	log.Printf("[%s] Name:%s 初始化成功", PackageNameSqlServer, c.config.DbName)
	return c
}

func (c *Component) initSqlServerDb() *gorm.DB {
	db, err := gorm.Open(sqlserver.Open(c.config.Dsn), &gorm.Config{})
	if err != nil {
		log.Printf("[%s] ❌ 连接异常 DB:%s, Err:%v", PackageNameSqlServer, c.config.DbName, err)
		return nil
	}
	idb, _ := db.DB()
	idb.SetMaxIdleConns(c.config.MaxIdleConns)
	idb.SetMaxOpenConns(c.config.MaxOpenConns)
	return db
}
