package idownload

import (
	"github.com/cute-angelia/go-xutils/syntax/ijson"
	"github.com/spf13/viper"
	"log"
	"runtime"
	"strings"
	"time"
)

type Option func(c *Container)

type Container struct {
	config *config
	name   string
}

func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
	}
}

// Load viper 加载 配置
func Load(key string) *Component {
	iconfig := DefaultConfig()
	configData := viper.GetStringMap(key)
	jsonstr, _ := ijson.Marshal(configData)
	if err := ijson.Unmarshal(jsonstr, &iconfig); err != nil {
		log.Println(err)
	}
	return newComponent(iconfig)
}

// New options 模式
func New(options ...Option) *Component {
	c := &Container{
		config: DefaultConfig(),
	}
	for _, option := range options {
		option(c)
	}
	return newComponent(c.config)
}

// WithProxyHttp   string // 代理 http://ip:port
func WithProxyHttp(ProxyHttp string) Option {
	return func(c *Container) {
		c.config.ProxyHttp = ProxyHttp
	}
}

// WithProxySocks5 string // 代理 ip:port
func WithProxySocks5(ProxySocks5 string) Option {
	ProxySocks5 = strings.Replace(ProxySocks5, "socks5://", "", -1)

	return func(c *Container) {
		c.config.ProxySocks5 = ProxySocks5
	}
}

// WithCookie  string // cookie
func WithCookie(Cookie string) Option {
	return func(c *Container) {
		c.config.Cookie = Cookie
	}
}

// WithUserAgent user-agent
func WithUserAgent(UserAgent string) Option {
	return func(c *Container) {
		c.config.UserAgent = UserAgent
	}
}

func WithHost(Host string) Option {
	return func(c *Container) {
		c.config.Host = Host
	}
}

func WithUserAgentRandon() Option {
	return func(c *Container) {
		c.config.UserAgent = RandomUserAgent()
	}
}

func WithUserAgentRandonMobile() Option {
	return func(c *Container) {
		c.config.UserAgent = RandomMobileUserAgent()
	}
}

func WithReferer(Referer string) Option {
	return func(c *Container) {
		c.config.Referer = Referer
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Container) {
		c.config.Timeout = timeout
	}
}

func WithDebug(Debug bool) Option {
	return func(c *Container) {
		c.config.Debug = Debug
	}
}

func WithProgressbar(Progressbar bool) Option {
	return func(c *Container) {
		c.config.Progressbar = Progressbar
	}
}

func WithResume(resume bool) Option {
	return func(c *Container) {
		c.config.Resume = resume
	}
}

func WithAuthorization(authorization string) Option {
	return func(c *Container) {
		c.config.Authorization = authorization
	}
}

func WithFileMax(fileMax int) Option {
	return func(c *Container) {
		c.config.FileMax = fileMax
	}
}

func WithConcurrency(concurrency int) Option {
	return func(c *Container) {
		c.config.Concurrency = concurrency
	}
}
func WithConcurrencyCpu() Option {
	concurrency := runtime.NumCPU()
	if concurrency >= 2 {
		concurrency = concurrency / 2
	}
	return func(c *Container) {
		c.config.Concurrency = concurrency
	}
}
func WithRetryAttempt(retryAttempt int) Option {
	return func(c *Container) {
		c.config.RetryAttempt = retryAttempt
	}
}
func WithRetryWaitTime(retryWaitTime time.Duration) Option {
	return func(c *Container) {
		c.config.RetryWaitTime = retryWaitTime
	}
}
