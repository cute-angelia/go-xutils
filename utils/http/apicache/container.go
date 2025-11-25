package apicache

import (
	"github.com/cute-angelia/go-utils/components/caches"
	"github.com/cute-angelia/go-xutils/utils/generator/hash"
	jsoniter "github.com/json-iterator/go"
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

func New(cache caches.Cache) *Container {
	c := DefaultContainer()
	c.config.Cache = cache
	return c
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Container) {
		c.config.Timeout = timeout
	}
}
func WithOnlyToday(onlyTodya bool) Option {
	return func(c *Container) {
		c.config.OnlyToday = onlyTodya
	}
}
func WithCacheKey(cacheKey string) Option {
	return func(c *Container) {
		c.config.CacheKey = cacheKey
	}
}

func WithGenerateCacheKey(params interface{}, filtes []string) Option {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	bparms, _ := json.Marshal(params)
	m := make(map[string]interface{})
	json.Unmarshal(bparms, &m)

	for _, filte := range filtes {
		if _, ok := m[filte]; ok {
			delete(m, filte)
		}
	}

	// log.Printf("%#v", m)

	cacheKey, _ := json.Marshal(m)
	cacheKeyMd5 := hash.NewEncodeMD5(string(cacheKey))

	return func(c *Container) {
		c.config.CacheKey = cacheKeyMd5
	}
}

func WithDebug(debug bool) Option {
	return func(c *Container) {
		c.config.Debug = debug
	}
}
func WithPrefix(prefix string) Option {
	return func(c *Container) {
		c.config.Prefix = prefix
	}
}

func WithPrefixMaxNum(prefixMaxNum int) Option {
	return func(c *Container) {
		c.config.PrefixMaxNum = prefixMaxNum
	}
}

// MustBuild  ...
func (c *Container) MustBuild(dbName string, prefix string, options ...Option) *Component {
	// Must
	c.config.Prefix = prefix

	for _, option := range options {
		option(c)
	}
	// log.Println(PackageName, fmt.Sprintf("%+v", c.config))

	if len(c.config.Prefix) == 0 {
		panic("cachekey prefix is empty")
	}
	if len(c.config.CacheKey) == 0 {
		// panic("cachekey is empty")
	}

	return newComponent(c.name, c.config)
}
