package igorm

import (
	"gorm.io/gorm/logger"
	"io"
	"time"
)

const packageName = "component.db.gorm"

// config options
type config struct {
	DbName       string
	DbFile       string // DbFile db数据存放路径
	Dsn          string
	MaxIdleConns int              // SetMaxIdleConns 用于设置连接池中空闲连接的最大数量(10)
	MaxOpenConns int              // SetMaxOpenConns 设置打开数据库连接的最大数量(100)
	MaxLifetime  time.Duration    // SetConnMaxLifetime 设置了连接可复用的最大时间。 time.Hour
	MaxIdleTime  time.Duration    // SetConnMaxIdleTime 设置连接最大空闲时间
	LoggerWriter io.Writer        // 外部 io.writer， 输出日志到外部
	LogLevel     logger.LogLevel  // 内部日志等级，GORM 定义了这些日志级别：Silent、Error、Warn、Info
	Logger       logger.Interface // 内部日志初始化,传递
	Debug        bool
}

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		MaxLifetime:  time.Hour,
		MaxIdleTime:  10 * time.Minute,
		LogLevel:     logger.Info,
	}
}
