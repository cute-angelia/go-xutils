package isingleflight

import (
	"golang.org/x/sync/singleflight"
	"time"
)

var gsf singleflight.Group

// GenericWrapper 使用泛型 T 封裝 singleflight
func GenericWrapper[T any](key string, duration time.Duration, fn func() (T, error)) (T, error) {
	val, err, _ := gsf.Do(key, func() (interface{}, error) {
		res, err := fn()
		if err != nil {
			return nil, err
		}
		// 保持 Key 被佔用以達到 1 秒內僅執行一次
		time.Sleep(duration)
		return res, nil
	})

	if err != nil {
		var zero T // 定義 T 的零值
		return zero, err
	}

	return val.(T), nil
}
