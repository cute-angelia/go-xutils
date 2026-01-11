package icmd

import (
	"github.com/go-cmd/cmd"
	"log"
	"time"
)

/*
package main

import (
	"fmt"

	"github.com/go-cmd/cmd"
)

func main() {
	// Create Cmd, buffered output
	envCmd := cmd.NewCmd("env")

	// Run and wait for Cmd to return Status
	status := <-envCmd.Start()

	// Print each line of STDOUT from Cmd
	for _, line := range status.Stdout {
		fmt.Println(line)
	}
}
*/
// Exec 执行外部命令并带超时控制
func Exec(exePath string, param []string, timeout time.Duration) cmd.Status {
	// 2026 建议：如果不需要捕获所有输出，建议在 cmd.Options 中关闭内存缓冲
	findCmd := cmd.NewCmd(exePath, param...)
	statusChan := findCmd.Start()

	// 1. 修正 Ticker 泄漏：使用 context 或显式停止
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop() // 确保函数退出时停止定时器

	done := make(chan struct{})

	// 2. 状态监控协程
	go func() {
		for {
			select {
			case <-ticker.C:
				status := findCmd.Status()
				n := len(status.Stdout)
				if n > 0 {
					log.Println("Current Last Line:", status.Stdout[n-1])
				}
			case <-done:
				return
			}
		}
	}()

	// 3. 超时控制
	var finalStatus cmd.Status
	select {
	case finalStatus = <-statusChan:
		// 命令正常结束
	case <-time.After(timeout):
		// 响应超时
		_ = findCmd.Stop()
		finalStatus = <-statusChan
	}

	// 4. 清理资源
	close(done)
	return finalStatus
}
