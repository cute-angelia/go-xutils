package ifile

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// Dir 获取目录路径
func Dir(fpath string) string {
	return filepath.Dir(fpath)
}

// Name 获取文件/目录名（带扩展名）
func Name(fpath string) string {
	return filepath.Base(fpath)
}

// NameNoExt 获取文件名（不包含扩展名）
func NameNoExt(filePath string) string {
	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)
	// 只有当存在扩展名时才截断
	if ext != "" {
		return fileName[:len(fileName)-len(ext)]
	}
	return fileName
}

// FileExt 获取文件后缀（专门处理 URL 场景，转为小写）
func FileExt(fpath string) string {
	// 如果是处理 URL，建议使用标准的 path 包
	if strings.Contains(fpath, "?") {
		fpath = strings.Split(fpath, "?")[0]
	}
	// URL 路径通常使用正斜杠，使用 path.Ext 比较安全
	return strings.ToLower(path.Ext(fpath))
}

// Suffix 获取系统文件后缀（保留原始大小写）
func Suffix(fpath string) string {
	// 处理本地文件路径建议使用 filepath
	return filepath.Ext(fpath)
}

// FilterFilename 增强版正则
func FilterFilename(name string) string {
	// 增加对 : * \ 等字符的过滤，确保跨平台安全
	regex := regexp.MustCompile(`[|&;$%@"<>()+,?:\s\x00-\x1f\\*/]`)
	return regex.ReplaceAllString(name, "-")
}

// GetHomeDir 获取当前用户家目录
// 2026 推荐做法：直接使用标准库 os.UserHomeDir()
func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// 如果获取失败（极少见），返回当前目录或临时目录
		return "./"
	}
	return home
}

// 如果你确实需要获取 XDG 配置目录（这与家目录不同）
func GetConfigDir() string {
	// os.UserConfigDir() 会在：
	// Linux 返回 $XDG_CONFIG_HOME 或 ~/.config
	// Windows 返回 %AppData%
	// macOS 返回 ~/Library/Application Support
	dir, err := os.UserConfigDir()
	if err != nil {
		return GetHomeDir()
	}
	return dir
}

// Mkdir 修正：移除对 dirPath 的正则过滤
// 注意：MkdirAll 的路径不应该被过滤，否则传入的 "C:/Data" 可能会变成 "CData" 导致创建失败
func Mkdir(dirPath string, perm os.FileMode) error {
	return os.MkdirAll(dirPath, perm)
}

// OpenFile 修正：统一使用 filepath
func OpenFile(fpath string, flag int, perm os.FileMode) (*os.File, error) {
	fileDir := filepath.Dir(fpath)
	if err := os.MkdirAll(fileDir, DefaultDirPerm); err != nil {
		return nil, err
	}
	return os.OpenFile(fpath, flag, perm)
}

// GetFileWithLocal 修正：使用 io.ReadAll 替代 ioutil
func GetFileWithLocal(fpath string) ([]byte, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// DeleteFile 修正：移除打印，增加错误返回
func DeleteFile(fpath string) error {
	err := os.Remove(fpath)
	if err != nil {
		return err
	}
	// 2026 生产标准：除非是 Debug 模式，否则不建议在库函数里直接 Println
	return nil
}
