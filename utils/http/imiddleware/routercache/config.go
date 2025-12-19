package routercache

import (
	"github.com/cute-angelia/go-xutils/components/caches"
	"net/http"
	"time"
)

const PackageName = "component.utils.http.routercache"

// config options
type config struct {
	Store caches.Cache  // 缓存器
	Ttl   time.Duration // 过期时间

	Methods          []string
	StatusCodeFilter func(int) bool // 特定code生效 默认小于400

	HeadReqRefreshKey string // 头部信息带刷新key 则刷新缓存
	HeadRespExpire    bool   // 头部返回过期时间

	CustomKey string // 自定义key

	PrintLog bool // 打印日志
}

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{
		Ttl:               time.Minute * 10,
		HeadReqRefreshKey: "xfresh",
		Methods:           []string{http.MethodPost, http.MethodGet},
		HeadRespExpire:    true,
		CustomKey:         "",
		PrintLog:          true,
	}
}
