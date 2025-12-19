package loggerV3

import (
	"github.com/cute-angelia/go-xutils/syntax/ifile"
	"github.com/cute-angelia/go-xutils/syntax/ijson"
	"github.com/rs/zerolog"
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

func WithProject(project string) Option {
	return func(c *Container) {
		c.config.Project = project
	}
}

func WithIsOnline(isOnline bool) Option {
	return func(c *Container) {
		c.config.IsOnline = isOnline
	}
}

func WithMaxSize(maxSize int) Option {
	return func(c *Container) {
		c.config.MaxSize = maxSize
	}
}
func WithMaxBackups(maxBackups int) Option {
	return func(c *Container) {
		c.config.MaxBackups = maxBackups
	}
}
func WithMaxAge(maxAge int) Option {
	return func(c *Container) {
		c.config.MaxAge = maxAge
	}
}
func WithEveryday(everyday bool) Option {
	return func(c *Container) {
		c.config.Everyday = everyday
	}
}

func WithLevel(level zerolog.Level) Option {
	return func(c *Container) {
		c.config.Level = level
	}
}

func WithFileJson(fileJson bool) Option {
	return func(c *Container) {
		c.config.FileJson = fileJson
	}
}

func WithHookError(hookError bool) Option {
	return func(c *Container) {
		c.config.HookError = hookError
	}
}

func WithLogPath(logPath string) Option {
	return func(c *Container) {
		ifile.MkParentDir(logPath)
		c.config.LogPath = logPath
	}
}
