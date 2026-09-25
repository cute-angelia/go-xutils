package apiV3

import (
	"context"
	"net/http"
)

// CryptoEr 从中间件设置加密密钥
var CryptoEr = cryptoEr{}

type cryptoEr struct {
}

var (
	CryptoCtxKey = &contextKey{"CryptoKey"}
)

func (that cryptoEr) SetCryptoKey(cryptoKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), CryptoCtxKey, cryptoKey))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// GetCryptoKey 从请求 Context 中读取由 SetCryptoKey 中间件注入的加密密钥。
// 若未注入则返回空字符串（表示不加密）。
func (that cryptoEr) GetCryptoKey(r *http.Request) string {
	if value, ok := r.Context().Value(CryptoCtxKey).(string); ok {
		return value
	}
	return ""
}
