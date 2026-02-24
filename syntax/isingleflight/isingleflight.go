package isingleflight

import (
	"time"

	"golang.org/x/sync/singleflight"
)

var gsf singleflight.Group

// GenericWrapper 使用泛型 T 封裝 singleflight
func GenericWrapper[T any](key string, duration time.Duration, fn func() (T, error)) (T, error) {
	val, err, _ := gsf.Do(key, func() (interface{}, error) {
		res, err := fn()
		if err != nil {
			return nil, err
		}
		// 注意：这会让第一个请求者变慢
		if duration > 0 {
			time.Sleep(duration)
		}
		return res, nil
	})

	var zero T
	if err != nil {
		return zero, err
	}

	// 安全断言
	if typedVal, ok := val.(T); ok {
		return typedVal, nil
	}
	return zero, nil
}
