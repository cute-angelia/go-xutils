package random

import (
	crand "crypto/rand"
	"math/rand/v2" // 2026 年推荐使用 math/rand/v2
)

// RandInt 生成 [min, max) 之间的随机整数
func RandInt(min, max int) int {
	if min == max {
		return min
	}
	if max < min {
		min, max = max, min
	}

	// 2026 推荐写法：使用 Go 1.22+ 的 rand.IntN
	// 它使用了全局高度优化的 ChaCha8 算法，并发性能极强，无需手动创建 Source
	return min + rand.IntN(max-min)
}

// RandBytes 生成随机字节切片（加密级安全）
func RandBytes(length int) []byte {
	if length < 1 {
		return []byte{}
	}
	b := make([]byte, length)

	// 使用 crypto/rand.Read 是最稳健的写法
	if _, err := crand.Read(b); err != nil {
		return nil
	}
	return b
}
