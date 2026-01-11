package ifile

import (
	"github.com/guonaihong/gout"
	"os"
	"path/filepath"
	"time"
)

// DownloadFileWithSrc 修正版：使用 BindBody 自动处理流向，避免 Unresolved reference
func DownloadFileWithSrc(src string, dir string, filename string) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	destPath := filepath.Join(dir, filename)
	out, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	// 注意：这里不要在 defer 里直接 Close 却不检查错误
	defer out.Close()

	// 2026 最佳实践：BindBody 可以直接接收 io.Writer，实现流式拷贝
	err = gout.GET(src).SetHeader(gout.H{
		"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}).
		SetTimeout(time.Second * 30).
		// 核心：直接绑定到文件句柄，数据会直接流向文件
		BindBody(out).
		F().
		Retry().
		Attempt(3).
		Do()

	if err != nil {
		out.Close()             // 手动关闭以便删除
		_ = os.Remove(destPath) // 下载失败清理残留文件
		return "", err
	}

	return destPath, out.Close()
}
