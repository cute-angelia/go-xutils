package imath

import (
	"math/rand"
)

// RandomInt 返回 [min, max) 范围内的随机整数
// 注意：max 必须大于 min
func RandomInt(min, max int) int {
	if min >= max {
		return min
	}
	// 2026 推荐写法：直接使用 rand.Intn，无需手动 Seed
	// Go 1.20+ 会自动处理种子初始化
	return min + rand.Intn(max-min)
}
