package ffmpeg

import (
	"github.com/cute-angelia/go-xutils/syntax/ifile"
	"log"
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

func Load() *Container {
	c := DefaultContainer()
	return c
}

func WithFfmpegPath(ffmpegPath string) Option {
	return func(c *Container) {
		if !ifile.IsExist(ffmpegPath) {
			log.Println("ffmpegPath not exist", ffmpegPath)
		}
		c.config.FfmpegPath = ffmpegPath
	}
}

func WithFilesPath(filesPath string) Option {
	return func(c *Container) {
		c.config.FilesPath = filesPath
	}
}

func WithTimeOut(timeout time.Duration) Option {
	return func(c *Container) {
		c.config.Timeout = timeout
	}
}

// Build ...
func (c *Container) Build(options ...Option) *Component {
	for _, option := range options {
		option(c)
	}
	return newComponent(c.name, c.config)
}
