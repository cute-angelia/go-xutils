package isingleflight

import (
	"golang.org/x/sync/singleflight"
)

var gsf singleflight.Group

// GenericWrapper 使用泛型 T 封裝 singleflight
func GenericWrapper[T any](key string, fn func() (T, error)) (T, error) {
	val, err, _ := gsf.Do(key, func() (interface{}, error) {
		res, err := fn()
		if err != nil {
			return nil, err
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
