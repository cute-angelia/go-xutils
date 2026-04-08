package idownload

import (
	"time"
)

const PackageName = "component.idownload"

// config options
type config struct {
	// 代理
	ProxyHttp   string // 代理 http://ip:port
	ProxySocks5 string // 代理 ip:port

	// 头部信息
	Cookie                   string // cookie
	Referer                  string // Referer
	Host                     string
	UserAgent                string // user-agent
	UseRandomUserAgent       bool
	UseRandomUserAgentMobile bool
	Authorization            string

	// 重试
	RetryAttempt  int           // 重试次数
	RetryWaitTime time.Duration // 重试间隔

	FileMax int

	// 并发数量
	Concurrency int
	// 分片下载
	Resume bool

	// 超时
	Timeout         time.Duration // // 单次 HTTP 请求超时
	DownloadTimeout time.Duration // 整体下载超时（新增）；为 0 时自动推算为 Timeout*(Concurrency+1)
	Debug           bool          //  debug 日志

	Progressbar bool // 进度条开关
}

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{
		Concurrency:              0,
		Resume:                   true,
		ProxyHttp:                "",
		ProxySocks5:              "",
		Cookie:                   "",
		UserAgent:                "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1",
		UseRandomUserAgent:       false,
		UseRandomUserAgentMobile: false,
		Referer:                  "",
		Debug:                    false,
		Authorization:            "",
		RetryAttempt:             3,
		RetryWaitTime:            time.Second * 5,
		Progressbar:              false,
		FileMax:                  -1,
	}
}
