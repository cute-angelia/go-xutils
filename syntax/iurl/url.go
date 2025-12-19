package iurl

import (
	"fmt"
	"github.com/cute-angelia/go-xutils/syntax/istrings"
	"net/url"
	"strings"
)

// GetDomainWithOutSlant 检查域名最后是否存在斜线并返回无斜线域名
func GetDomainWithOutSlant(domain string) string {
	// 是否为空
	if domain == "" {
		return ""
	}

	// 获取最后一个字符
	last := domain[len(domain)-1:]
	// 如果是斜线
	if last == "/" {
		domain = domain[:len(domain)-1]
	}

	return domain
}

// CleanUrlWithoutParam 获得一个干净的不带参数的地址
func CleanUrlWithoutParam(uri string) string {
	clearUrl, _ := url.QueryUnescape(uri)
	if uriInfo, err := url.Parse(clearUrl); err == nil {
		if len(uriInfo.Scheme) == 0 {
			uriInfo.Scheme = "https"
		}
		return fmt.Sprintf("%s://%s%s", uriInfo.Scheme, uriInfo.Host, uriInfo.Path)
	} else {
		if beforeUrl, _, ok := istrings.Cut(clearUrl, `?`); ok {
			if !strings.Contains(beforeUrl, "http") {
				beforeUrl = "https:" + beforeUrl
			}
			return beforeUrl
		} else {
			return clearUrl
		}
	}
}

// RemoveParam 移除Query指定字段
func RemoveParam(uri string, removes []string) (string, error) {
	if u, err := url.Parse(uri); err != nil {
		return uri, err
	} else {
		m, _ := url.ParseQuery(u.RawQuery)
		// 移除
		for k, _ := range m {
			for _, v := range removes {
				if k == v {
					delete(m, k)
				}
			}
		}
		return fmt.Sprintf("%s://%s%s?%s", u.Scheme, u.Host, u.Path, m.Encode()), nil
	}
}

/*
Encode 文本编码
one ->

	url.QueryEscape(text)

multi ->

	params := url.Values{}
	params.Add("q", "1 + 2")
	params.Add("s", "example for GoLang.com")
	output := params.Encode()
*/
func Encode(s string) string {
	return url.QueryEscape(s)
}

// EncodeQuery Url Query 编码
func EncodeQuery(uri string) string {
	if strings.Contains(uri, "http") {
		if l2, err := url.Parse(uri); err != nil {
			return uri
		} else {
			return fmt.Sprintf("%s://%s%s?%s", l2.Scheme, l2.Host, l2.Path, l2.Query().Encode())
		}
	} else {
		if l, err := url.ParseQuery(uri); err != nil {
			return uri
		} else {
			return l.Encode()
		}
	}
}

// Decode 解码文本
// 如果文本是URL，请用 url.Parse |  url.PathUnescape(path) | url.ParseQuery(queryStr)
func Decode(s string) string {
	if d, err := url.QueryUnescape(s); err != nil {
		return s
	} else {
		return d
	}
}
