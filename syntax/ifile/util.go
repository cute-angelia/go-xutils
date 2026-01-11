package ifile

import (
	"io"
	"net/http"
	"os"
)

const (
	// MimeSniffLen 标准嗅探长度
	MimeSniffLen = 512
)

// MimeType 修正：确保文件句柄被正确复位并安全关闭
func MimeType(path string) (mime string) {
	if path == "" {
		return ""
	}

	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	// 修正：确保在函数退出时关闭文件
	defer file.Close()

	return ReaderMimeType(file)
}

// ReaderMimeType 修正：处理 io.ReadSeeker 以防数据读取后导致流丢失
func ReaderMimeType(r io.Reader) (mime string) {
	// 定义嗅探缓冲区
	buf := make([]byte, MimeSniffLen)

	// 使用 io.ReadAtLeast 替代 ReadFull，防止小文件报错
	n, err := io.ReadAtLeast(r, buf, 1)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "application/octet-stream"
	}

	if n == 0 {
		return "application/octet-stream"
	}

	// 核心修正：如果 r 支持 Seek (如 *os.File)，必须将指针移回开头
	// 否则后续读取该 Reader 会丢失前 n 字节
	if rs, ok := r.(io.Seeker); ok {
		_, _ = rs.Seek(0, io.SeekStart)
	}

	return http.DetectContentType(buf[:n])
}
