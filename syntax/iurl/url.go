package iurl

import (
	"net/url"
	"strings"
)

// GetDomainWithOutSlant 移除域名末尾的斜杠
func GetDomainWithOutSlant(domain string) string {
	// 2026 推荐做法：使用内置的高效函数
	return strings.TrimSuffix(domain, "/")
}

// CleanUrlWithoutParam 获得一个干净的不带参数的地址
func CleanUrlWithoutParam(uri string) string {
	if uri == "" {
		return ""
	}

	// 优先尝试标准解析
	u, err := url.Parse(uri)
	if err != nil {
		// 解析失败时，回退到字符串截断逻辑
		before, _, _ := strings.Cut(uri, "?")
		if !strings.Contains(before, "://") && strings.Contains(before, ".") {
			before = "https://" + strings.TrimPrefix(before, "//")
		}
		return before
	}

	// 补全 Scheme
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	// 清空 Query 和 Fragment 获得纯净 URL
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

// RemoveParam 移除 Query 指定字段（支持保留 Fragment）
func RemoveParam(uri string, removes []string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return uri, err
	}

	values := u.Query()
	for _, key := range removes {
		values.Del(key) // 2026 标准：values.Del 比循环遍历更高效
	}

	u.RawQuery = values.Encode()
	return u.String(), nil
}

// Encode 文本编码（Query 参数值编码）
func Encode(s string) string {
	return url.QueryEscape(s)
}

// EncodeQuery 对 URL 的 Query 部分进行规范化重新编码
func EncodeQuery(uri string) string {
	if !strings.Contains(uri, "://") && !strings.Contains(uri, "?") {
		// 处理纯参数字符串: a=1&b=2
		if l, err := url.ParseQuery(uri); err == nil {
			return l.Encode()
		}
		return uri
	}

	u, err := url.Parse(uri)
	if err != nil {
		return uri
	}

	// 重新编码 Query 并保留 URL 其他部分
	u.RawQuery = u.Query().Encode()
	return u.String()
}

// Decode 解码文本
func Decode(s string) string {
	if d, err := url.QueryUnescape(s); err == nil {
		return d
	}
	return s
}
