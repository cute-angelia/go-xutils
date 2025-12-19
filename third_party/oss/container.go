package oss

import (
	"fmt"
	"github.com/cute-angelia/go-xutils/syntax/ijson"
	"github.com/spf13/viper"
	"log"
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

func WithAccessKeyId(AccessKeyId string) Option {
	return func(c *Container) {
		c.config.AccessKeyId = AccessKeyId
	}
}
func WithAccessKeySecret(AccessKeySecret string) Option {
	return func(c *Container) {
		c.config.AccessKeySecret = AccessKeySecret
	}
}
func WithEndpoint(Endpoint string) Option {
	return func(c *Container) {
		c.config.Endpoint = Endpoint
	}
}
func WithBucketName(BucketName string) Option {
	return func(c *Container) {
		c.config.BucketName = BucketName
	}
}
func WithBucketHost(BucketHost string) Option {
	return func(c *Container) {
		c.config.BucketHost = BucketHost
	}
}
func WithDebug(Debug bool) Option {
	return func(c *Container) {
		c.config.Debug = Debug
	}
}

// Build ...
func (c *Container) MustBuild(options ...Option) *Component {
	for _, option := range options {
		option(c)
	}

	if len(c.config.BucketName) == 0 ||
		len(c.config.AccessKeyId) == 0 ||
		len(c.config.AccessKeySecret) == 0 {
		log.Println(PackageName, fmt.Sprintf("ERROR: %+v", c.config))
	}
	return newComponent(c.config)
}
