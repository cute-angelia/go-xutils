package ipath

import (
	"path"
	"path/filepath"
	"strings"
)

// Clean 专门处理 URL 风格路径，去除参数并规范化
func Clean(p string) string {
	// Go 1.18+ 推荐用法
	v, _, _ := strings.Cut(p, "?")
	// 统一转为正斜杠处理，并去除首尾斜杠
	return strings.Trim(path.Clean("/"+v), "/")
}

// GetFileName 获取文件名（不含扩展名）
func GetFileName(filePath string) string {
	filePath = Clean(filePath)
	if filePath == "." || filePath == "" {
		return ""
	}
	base := path.Base(filePath)
	ext := path.Ext(base)
	return base[:len(base)-len(ext)]
}

// GetFileNameAndExt 获取文件名和后缀（后缀转小写）
func GetFileNameAndExt(filePath string) (name string, ext string) {
	filePath = Clean(filePath)
	if filePath == "." || filePath == "" {
		return "", ""
	}
	base := path.Base(filePath)
	ext = strings.ToLower(path.Ext(base))
	name = base[:len(base)-len(ext)]
	return
}

// GetFileExt 获取后缀（带点，如 .jpg）
func GetFileExt(filePath string) (ext string, found bool) {
	filePath = Clean(filePath)
	ext = strings.ToLower(path.Ext(filePath))
	return ext, ext != ""
}

// OSPathClean 如果你需要处理本地文件系统路径，请调用此函数
func OSPathClean(p string) string {
	return filepath.Clean(p)
}
