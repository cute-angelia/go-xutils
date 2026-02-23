package ifile

import (
	"fmt"
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

// 1. 覆盖写模式 (Truncate/Overwrite)
// 适用于日志重置、配置文件更新等。如果文件不存在则创建，如果存在则清空。
// Flag: os.O_RDWR | os.O_CREATE | os.O_TRUNC
// Perm: 0666 (文件), 0755 (目录)
func CreateFile(fpath string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(fpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// 2. 追加写模式 (Append)
func AppendFile(fpath string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(fpath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
}

// 3. 只读模式 (Read Only)
func OpenReadOnly(fpath string) (*os.File, error) {
	return os.Open(fpath) // 内部封装了 os.OpenFile(name, O_RDONLY, 0)
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

// RenamePath 直接重命名（支持文件和文件夹）
func RenamePath(src, dst string) error {
	// 1. 检查原路径是否存在
	if _, err := os.Stat(src); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("源路径不存在: %s", src)
		}
		return fmt.Errorf("检查源路径时出错: %v", err)
	}
	// 2. 直接重命名
	// 注意：如果 src 和 dst 不在同一个分区，这里会报错 "invalid cross-device link"
	err := os.Rename(src, dst)
	if err != nil {
		return fmt.Errorf("重命名失败: %v", err)
	}
	return nil
}

// MvCrossDevice 移動檔案，具備跨裝置/分區的相容性
func MvCrossDevice(src, dst string) error {
	// 1. 首先嘗試使用原生的 os.Rename (效能最高)
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// 2. 如果發生跨分區錯誤 (invalid cross-device link)，則改用 Copy + Delete
	// 這裡不直接返回錯誤，而是進行手動拷貝
	return moveCrossDevice(src, dst)
}

func moveCrossDevice(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// 獲取原檔案權限
	info, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	// 建立目標檔案
	destFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	// 拷貝內容
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// 關閉檔案後再刪除原檔案
	sourceFile.Close()
	return os.Remove(src)
}
