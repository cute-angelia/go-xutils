package ifileutil

import (
	"github.com/cute-angelia/go-xutils/syntax/ijson"
	"github.com/spf13/viper"
	"log"
)

type Option func(c *Container)

// Container 两种方式
// 一种从 viper 获取配置
// 一种是 options 模式
type Container struct {
	config *config
}

// Load viper 加载 配置
func Load(key string) *Component {
	iconfig := DefaultConfig()
	configData := viper.GetStringMap(key)
	jsonstr, _ := ijson.Marshal(configData)
	if err := ijson.Unmarshal(jsonstr, &iconfig); err != nil {
		log.Println(err)
	}
	// log.Println(ijson.Pretty(iconfig))
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

func WithDir(dir string) Option {
	return func(c *Container) {
		c.config.Dir = dir
	}
}

func WithDirInclude(dirInclude []string) Option {
	return func(c *Container) {
		c.config.DirInclude = dirInclude
	}
}
func WithDirDeclude(dirDeclude []string) Option {
	return func(c *Container) {
		c.config.DirDeclude = dirDeclude
	}
}
func WithExtInclude(extInclude []string) Option {
	return func(c *Container) {
		c.config.ExtInclude = extInclude
	}
}
func WithExtDeclude(extDeclude []string) Option {
	return func(c *Container) {
		c.config.ExtDeclude = extDeclude
	}
}
