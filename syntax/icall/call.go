package icall

import (
	"log"
	"runtime"
	"runtime/debug"
)

// Call 执行函数并捕获其可能产生的 panic，确保程序不崩溃
func Call(fn func()) {
	if fn == nil {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			// 获取发生 panic 时的完整堆栈信息
			stack := debug.Stack()

			// 2026 实践：将错误和堆栈一同记录，但不重新触发 Panic
			log.Printf("[PANIC RECOVER] error: %v\nStack: %s", err, stack)

			// 如果你的项目有 Sentry 或日志中心，应在此处上报
			// reportToSentry(err, stack)
		}
	}()

	fn()
}

// Go 安全地开启一个新的协程
func Go(fn func()) {
	// 2026 建议：如果任务非常重要，可以在此处注入 Context 以便管理
	go Call(fn)
}

// CallWithErr 如果需要返回错误信息
func CallWithErr(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PANIC RECOVER] %v\n%s", r, debug.Stack())
			// 将 panic 包装成普通的 error 返回
			err = runtime.Error(r.(runtime.Error))
		}
	}()
	return fn()
}
