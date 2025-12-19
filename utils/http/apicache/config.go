package apicache

import (
	"github.com/cute-angelia/go-xutils/components/caches"
	"time"
)

const PackageName = "component.http.apicache"

// config options
// 缓存逻辑说明：
// 正常按 key：value 保存
// 但有些业务逻辑，有很多参数， 如基础分页，如果带有分页的参数的 cachekey 将统一保存到 "config.前缀" 下， 用于一次性更新缓存
type config struct {
	Timeout time.Duration // 正常失效时间

	Prefix       string // 缓存前缀
	PrefixMaxNum int    // 最大子类个数，LRU 概念 默认 100

	CacheKey string // 保存的 key

	OnlyToday bool // 是否凌晨刷新
	Debug     bool

	Cache caches.Cache
}

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{
		Timeout:      time.Minute * 10,
		OnlyToday:    false,
		Debug:        false,
		PrefixMaxNum: 100,
	}
}
