package orderId

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand/v2" // 2026 推荐使用 v2
	"strconv"
	"time"
)

// GenerateOrderId 优化版：[时间戳] + [6位随机数]
func GenerateOrderId() string {
	// 使用缓存的 format 提高性能
	timenow := time.Now().Format("20060102150405")

	// 2026 推荐：直接使用 rand.IntN(v2)，无需 Seed，并发安全且高性能
	return timenow + fmt.Sprintf("%06d", rand.IntN(1000000))
}

// GenerateTradeId 优化版：更强的唯一性保证
func GenerateTradeId() string {
	// 组合：Unix微秒 + 随机数 + 自增位（或更强的随机）
	now := time.Now()
	str := strconv.FormatInt(now.UnixMicro(), 10) + strconv.Itoa(rand.IntN(999999))

	// 如果依然需要 MD5 格式的 ID (32位字符串)
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
