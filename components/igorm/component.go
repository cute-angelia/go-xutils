package igorm

import "sync"

// grom pool
var gormPool sync.Map

// 定义全局的 sync.Map 用于存储不同名称的 SQLite 数据库连接实例
var gormPoolSQLite sync.Map

var gormPoolPostgreSQL sync.Map // 用于存储不同名称的 PostgreSQL 数据库连接实例

type Component struct {
	config *config
	locker sync.Mutex
}

// newComponent ...
func newComponent(config *config) *Component {
	return &Component{
		config: config,
	}
}
