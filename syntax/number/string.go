package random

import (
	"math/rand/v2" // 2026 年推荐使用 v2 包
	"unsafe"
)

type Letter string

const (
	LetterAbc            = Letter("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	LetterAbcLower       = Letter("abcdefghijklmnopqrstuvwxyz")
	LetterAbcUpper       = Letter("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	LetterNumAndLowAbc   = Letter("abcdefghijklmnopqrstuvwxyz0123456789")
	LetterNumAndUpperAbc = Letter("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	LetterAll            = Letter("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

// RandString 2026 优化版
func RandString(length int, letters Letter) string {
	if length <= 0 {
		return ""
	}

	b := make([]byte, length)
	l := uint32(len(letters))

	for i := 0; i < length; i++ {
		// 2026 推荐：直接使用 rand.Uint32N，底层使用高性能 ChaCha8 算法
		// 无需手动创建 rand.NewSource，并发安全且零锁竞争
		b[i] = letters[rand.Uint32N(l)]
	}

	// 使用 unsafe 方式转换字符串（Go 1.22+ 常见优化，减少一次内存拷贝）
	return unsafe.String(&b[0], length)
}
