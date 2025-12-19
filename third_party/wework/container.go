package wework

import (
	"github.com/cute-angelia/go-xutils/syntax/ijson"
	"github.com/go-redis/redis/v8"
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

func WithCorpid(corpid string) Option {
	return func(c *Container) {
		c.config.CorpId = corpid
	}
}

func WithAgentId(agentId string) Option {
	return func(c *Container) {
		c.config.AgentId = agentId
	}
}
func WithSecret(secret string) Option {
	return func(c *Container) {
		c.config.Secret = secret
	}
}

func WithRedisClient(redisClient *redis.Client) Option {
	return func(c *Container) {
		c.config.redisClient = redisClient
	}
}
