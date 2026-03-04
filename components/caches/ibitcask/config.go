package ibitcask

import (
	"time"
)

// config options
type config struct {
	Path            string
	MaxDatafileSize int
	GcInterval      time.Duration

	MaxKeySize   uint32 // 允许 Key 达到 1KB (根据需求调整)
	MaxValueSize uint64 // 允许 Value 达到 10MB (根据需求调整)

}

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{
		Path:            "./bitcask",
		MaxDatafileSize: 1024 * 1024 * 200,
		GcInterval:      time.Hour,
		MaxKeySize:      1024,     // 允许 Key 达到 1KB
		MaxValueSize:    10 << 20, // 10MB
	}
}
