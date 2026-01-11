package ifile

import (
	"bytes"
	"os"
	"path/filepath" // 统一使用 filepath 处理系统路径
	"strings"
)

var (
	DefaultDirPerm  os.FileMode = 0775
	DefaultFilePerm os.FileMode = 0664
)

const (
	// CheckTypeExt 通过文件扩展名检查 (如 .jpg)
	CheckTypeExt int32 = iota
	// CheckTypeMimeType 通过文件内容 MimeType 检查 (如 image/jpeg)
	CheckTypeMimeType
)

// IsExist 现代写法
func IsExist(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// IsDir 报告路径是否为目录
func IsDir(path string) bool {
	if fi, err := os.Stat(path); err == nil {
		return fi.IsDir()
	}
	return false
}

// CheckIsEmptyDir 检查是否为空文件夹
func CheckIsEmptyDir(dirpath string) bool {
	files, err := os.ReadDir(dirpath)
	if err != nil {
		return false
	}
	for _, file := range files {
		if file.Name() != ".DS_Store" {
			return false
		}
	}
	return true
}

// IsFile 报告路径是否为文件
func IsFile(path string) bool {
	if fi, err := os.Stat(path); err == nil {
		return !fi.IsDir()
	}
	return false
}

// IsAbsPath 跨平台绝对路径判断
func IsAbsPath(aPath string) bool {
	return filepath.IsAbs(aPath)
}

// ImageMimeTypes 修正：增加常用映射
var ImageMimeTypes = map[string]string{
	"bmp":  "image/bmp",
	"gif":  "image/gif",
	"jpg":  "image/jpeg",
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"svg":  "image/svg+xml",
	"ico":  "image/x-icon",
	"webp": "image/webp",
	"avif": "image/avif", // 2026年常用格式
}

// IsImage 修正大小写敏感及逻辑冗余
func IsImage(uri string, checkType int32) bool {
	if checkType == CheckTypeExt {
		// 假设 NewFileName 已修正为使用 filepath
		uri = NewFileName(uri).CleanUrl()
		ext := strings.ToLower(filepath.Ext(uri))
		if len(ext) > 0 {
			ext = ext[1:] // 去掉点
		}
		_, ok := ImageMimeTypes[ext]
		return ok
	}

	if checkType == CheckTypeMimeType {
		mime := MimeType(uri) // 假设 MimeType 是你实现的另一个函数
		for _, imgMime := range ImageMimeTypes {
			if imgMime == mime {
				return true
			}
		}
	}
	return false
}

// IsZip 增加资源释放保护
func IsZip(filepath string) bool {
	f, err := os.Open(filepath)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 4)
	if n, err := f.Read(buf); err != nil || n < 4 {
		return false
	}
	return bytes.Equal(buf, []byte{0x50, 0x4B, 0x03, 0x04}) // 推荐使用 hex 字面量
}

// IsCanWrite 修正：文件不存在应返回 false
func IsCanWrite(inFile string) bool {
	file, err := os.OpenFile(inFile, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// IsCanRead 同上
func IsCanRead(inFile string) bool {
	file, err := os.OpenFile(inFile, os.O_RDONLY, 0)
	if err != nil {
		return false
	}
	file.Close()
	return true
}
