package apiV3

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/stampede"
)

var Stampeder = stampeder{}

type stampeder struct{}

// CachedPost 优化版：增加了内存保护和性能优化
func (s stampeder) CachedPost(cacheSize int, ttl time.Duration) func(next http.Handler) http.Handler {
	return stampede.HandlerWithKey(cacheSize, ttl, func(r *http.Request) uint64 {
		var buf []byte

		if r.Body != nil {
			// 1. 2026 安全实践：限制读取的 Body 大小，防止恶意内存攻击
			// 这里的 1MB (1<<20) 是针对 API 请求的常规限制
			lr := io.LimitReader(r.Body, 1<<20)

			// 2. 读取限制内的数据
			readBuf, err := io.ReadAll(lr)
			if err == nil {
				buf = readBuf
				// 3. 复写 Body 指针以便后续 Handler (如 Decode) 能再次读取
				// 使用 bytes.NewReader 比 bytes.NewBuffer 在只读场景更高效
				r.Body = io.NopCloser(bytes.NewReader(buf))
			}
		}

		// 4. 2026 性能优化：尽量减少 strings.ToLower 分配
		token := r.Header.Get("Authorization")

		// 结合 Path, Authorization 和 Body 计算唯一缓存 Key
		return stampede.BytesToHash(
			[]byte(r.URL.Path), // 如果确定路由区分大小写，可省去 ToLower
			[]byte(token),
			buf,
		)
	})
}
