package loggerV3

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	component   *Component
	mu          sync.RWMutex
	loggerError *zerolog.Logger
)

type Component struct {
	config *config
	logger *zerolog.Logger
	cancel context.CancelFunc // 用于停止每日轮转协程
	once   sync.Once          // 每个实例独立的初始化锁
}

// GetLogger 获取单例
func GetLogger() *zerolog.Logger {
	mu.RLock()
	if component != nil && component.logger != nil {
		defer mu.RUnlock()
		return component.logger
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()
	// 双重检查
	if component == nil {
		fmt.Println("日志未初始化，启动默认配置x")
		component = NewComponent(DefaultConfig())
	}
	return component.logger
}

// NewComponent 创建新组件（支持外部显式初始化）
func NewComponent(config *config) *Component {
	ctx, cancel := context.WithCancel(context.Background())
	cpt := &Component{
		config: config,
		cancel: cancel,
	}

	// 重点：先加锁赋值，让 GetLogger() 能拿到实例，不再触发“默认初始化”
	mu.Lock()
	component = cpt
	mu.Unlock()

	// 然后再初始化内部的 zerolog 实例
	cpt.initLogger(ctx)

	log.Println(ComponentName + "日志启动成功")

	return cpt
}

// Stop 停止相关资源（如轮转协程）
func (self *Component) Stop() {
	if self.cancel != nil {
		self.cancel()
	}
}

// Stop 停止日志组件相关的后台协程（如每日切割协程）
func Stop() {
	mu.Lock()
	defer mu.Unlock()
	if component != nil {
		component.Stop() // 调用之前定义的 Component.Stop()
	}
}

func (self *Component) initLogger(ctx context.Context) {
	self.once.Do(func() {
		// 1. 初始化主 Logger
		mainLogName := "log_" + self.config.Project + ".log"
		ilog := self.makeLogger(ctx, mainLogName, false)

		// 2. 挂载 Error Hook
		if self.config.HookError {
			ilog = ilog.Hook(ErrorHook{})
			errLogName := filepath.Join("error", "log_error_"+self.config.Project+".log")
			errLogger := self.makeLogger(ctx, errLogName, true)
			loggerError = &errLogger
		}

		// 3. 配置全局 Zerolog 属性
		zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"
		zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
			return filepath.Base(file) + ":" + strconv.Itoa(line)
		}
		zerolog.SetGlobalLevel(self.config.Level)

		self.logger = &ilog
	})
}

/*
// 重点：使用 CallerWithSkipFrameCount(3)
// 3 层深度：
// [0] zerolog 内部
// [1] zerolog 封装
// [2] loggerV3 内部 (GetLogger)
// [3] 你的业务代码 (main.go)
*/
func (self *Component) makeLogger(ctx context.Context, logName string, isErrorStream bool) zerolog.Logger {
	var writer io.Writer

	// 1. 确定输出流
	if self.config.IsOnline {
		rolling := self.newRollingFile(ctx, logName)
		if self.config.FileJson {
			writer = rolling
		} else {
			writer = self.formatLogger(rolling)
		}
	} else {
		writer = self.formatLogger(os.Stdout)
	}

	// 2. 配置原生 log 包 (标准库)
	// 关闭原生 log 的所有自带属性，因为 zerolog 会提供这些
	log.SetFlags(0)

	// 为原生 log 创建一个专用的 writer，Skip 设为 4
	// 这样 log.Println("xxx") 就会显示业务代码行号
	stdLogWriter := zerolog.New(writer).With().Timestamp().CallerWithSkipFrameCount(4).Logger()
	log.SetOutput(stdLogWriter)

	// 3. 业务使用的 Logger，Skip 设为 2
	// 链路：业务代码 -> GetLogger() -> zerolog.Info()
	l := zerolog.New(writer).With().Timestamp().CallerWithSkipFrameCount(2).Logger()

	// 4. 【关键】初始化成功提示不要用被劫持的 log.Println
	// 直接用 fmt 输出到控制台，或者用刚生成的 l 输出一次
	if !isErrorStream {
		fmt.Printf("[%s] 初始化成功: 项目=%s, 路径=%s\n", ComponentName, self.config.Project, self.config.LogPath)
	}

	return l
}

func (self *Component) formatLogger(out io.Writer) io.Writer {
	output := zerolog.ConsoleWriter{Out: out, TimeFormat: "2006-01-02 15:04:05.000"}
	output.FormatLevel = func(i interface{}) string {
		var l string
		if i == nil {
			l = "INFO" // 默认级别
		} else {
			l = strings.ToUpper(fmt.Sprintf("%s", i))
		}
		return fmt.Sprintf("| %-6s|", l)
	}
	return output
}

func (self *Component) newRollingFile(ctx context.Context, logName string) io.Writer {
	fullPath := filepath.Join(self.config.LogPath, logName)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	_ = os.MkdirAll(dir, 0755)

	ljLogger := &lumberjack.Logger{
		Filename:   fullPath,
		MaxBackups: self.config.MaxBackups,
		MaxSize:    self.config.MaxSize,
		MaxAge:     self.config.MaxAge,
		LocalTime:  true,
	}

	if self.config.Everyday {
		go func() {
			for {
				now := time.Now()
				next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
				timer := time.NewTimer(next.Sub(now))

				select {
				case <-timer.C:
					_ = ljLogger.Rotate()
				case <-ctx.Done():
					timer.Stop()
					return // 安全退出协程
				}
			}
		}()
	}

	return ljLogger
}
