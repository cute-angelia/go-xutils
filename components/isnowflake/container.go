package isnowflake

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

func WithMethod(method uint16) Option {
	return func(c *Container) {
		c.config.Method = method
	}
}

func WithBaseTime(baseTime int64) Option {
	return func(c *Container) {
		c.config.BaseTime = baseTime
	}
}

func WithWorkerId(WorkerId uint16) Option {
	return func(c *Container) {
		c.config.WorkerId = WorkerId
	}
}

func WithWorkerIdBitLength(WorkerIdBitLength byte) Option {
	return func(c *Container) {
		c.config.WorkerIdBitLength = WorkerIdBitLength
	}
}

func WithSeqBitLength(SeqBitLength byte) Option {
	return func(c *Container) {
		c.config.SeqBitLength = SeqBitLength
	}
}

func WithMaxSeqNumber(MaxSeqNumber uint32) Option {
	return func(c *Container) {
		c.config.MaxSeqNumber = MaxSeqNumber
	}
}

func WithMinSeqNumber(MinSeqNumber uint32) Option {
	return func(c *Container) {
		c.config.MinSeqNumber = MinSeqNumber
	}
}

func WithTopOverCostCount(TopOverCostCount uint32) Option {
	return func(c *Container) {
		c.config.TopOverCostCount = TopOverCostCount
	}
}
