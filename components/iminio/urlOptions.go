package iminio

import (
	"context"
	"github.com/cute-angelia/go-xutils/components/caches"
	"time"
)

type UrlOptions struct {
	Expiry  time.Duration // 签名过期时间（若为0则视为公共读）
	Version string        // 版本号（解决图片缓存刷新问题）
	Cache   caches.Cache  // 缓存对象
	Context context.Context
	Rebuild bool
}

type UrlOption func(*UrlOptions)

// WithExpiry 设置签名有效期（不设置则为公共拼接）
func WithExpiry(d time.Duration) UrlOption {
	return func(o *UrlOptions) { o.Expiry = d }
}

// WithVersion 设置版本号（如时间戳，强制刷新浏览器缓存）
func WithVersion(v string) UrlOption {
	return func(o *UrlOptions) { o.Version = v }
}

// WithCache 设置缓存
func WithCache(c caches.Cache) UrlOption {
	return func(o *UrlOptions) { o.Cache = c }
}

// WithContext 设置上下文
func WithContext(ctx context.Context) UrlOption {
	return func(o *UrlOptions) { o.Context = ctx }
}

func WithRebuild(rebuild bool) UrlOption {
	return func(o *UrlOptions) { o.Rebuild = rebuild }
}
