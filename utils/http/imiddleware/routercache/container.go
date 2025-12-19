package routercache

import (
	"github.com/cute-angelia/go-xutils/components/caches"
	"time"
)

type Option func(c *Container)

type Container struct {
	config *config
}

func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
	}
}

// New options 模式
func New(options ...Option) *Component {
	c := &Container{
		config: DefaultConfig(),
	}
	for _, option := range options {
		option(c)
	}

	if c.config.StatusCodeFilter == nil {
		c.config.StatusCodeFilter = func(code int) bool { return code < 400 }
	}

	if c.config.Store == nil {
		panic(PackageName + "need store")
	}

	return newComponent(c.config)
}

// WithTtl 过期时间
func WithTtl(ttl time.Duration) Option {
	return func(c *Container) {
		c.config.Ttl = ttl
	}
}

func WithStore(store caches.Cache) Option {
	return func(c *Container) {
		c.config.Store = store
	}
}

func WithHeadReqRefreshKey(refreshKey string) Option {
	return func(c *Container) {
		c.config.HeadReqRefreshKey = refreshKey
	}
}

func WithMethods(methods []string) Option {
	return func(c *Container) {
		c.config.Methods = methods
	}
}

func WithStatusCodeFilter(a func(int) bool) Option {
	return func(c *Container) {
		c.config.StatusCodeFilter = a
	}
}

func WithHeadRespExpire(headRespExpire bool) Option {
	return func(c *Container) {
		c.config.HeadRespExpire = headRespExpire
	}
}

func WithCustomKey(customKey string) Option {
	return func(c *Container) {
		c.config.CustomKey = customKey
	}
}

func WithPrintLog(printLog bool) Option {
	return func(c *Container) {
		c.config.PrintLog = printLog
	}
}
