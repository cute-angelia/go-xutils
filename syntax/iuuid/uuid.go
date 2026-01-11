package iuuid

import (
	crand "crypto/rand"
	"encoding/hex"
	"io"
)

// UUIdV4 生成符合 RFC 4122 标准的 UUID v4
func UUIdV4() (string, error) {
	var uuid [16]byte // 2026 优化：使用栈上的数组代替 make 切片，减少内存逃逸

	if _, err := io.ReadFull(crand.Reader, uuid[:]); err != nil {
		return "", err
	}

	// 设置版本号 (4) 和变体 (RFC 4122)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	// 2026 优化：使用固定长度缓冲区一次性构建字符串，避免 fmt.Sprintf
	var buf [36]byte
	hex.Encode(buf[0:8], uuid[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], uuid[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], uuid[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], uuid[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], uuid[10:])

	return string(buf[:]), nil
}
