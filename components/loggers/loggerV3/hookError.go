package loggerV3

import (
	"github.com/rs/zerolog"
)

type ErrorHook struct{}

func (h ErrorHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	// 仅处理 Error, Fatal, Panic 级别的日志
	if level >= zerolog.ErrorLevel {
		if loggerError != nil {
			// 将当前事件的元数据（如错误信息、堆栈等）写入错误日志文件
			// 注意：zerolog 的 Hook 是同步执行的
			loggerError.WithLevel(level).Msg(msg)
		}
	}
}
