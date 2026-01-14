package apiV3

import (
	"context"
	"net/http"
	"strings"
)

var ContentTyper = contentTyper{}

type contentTyper struct {
}

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation. This technique
// for defining context keys was copied from Go 1.7's new use of context in net/http.
type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "chi render context value " + k.name
}

var (
	ContentTypeCtxKey = &contextKey{"ContentType"}
)

// TypeContent is an enumeration of common HTTP content types.
type TypeContent int

// ContentTypes handled by this package.
const (
	ContentTypeUnknown TypeContent = iota
	ContentTypePlainText
	ContentTypeHTML
	ContentTypeJSON
	ContentTypeXML
	ContentTypeForm
	ContentTypeEventStream
	ContentTypeMultipart // 文件上传 multipart/form-data
)

func GetContentType(s string) TypeContent {
	// 1. 全部轉為小寫並去除兩端空格
	s = strings.ToLower(strings.TrimSpace(s))

	// 2. 取分號前的主體部分
	if parts := strings.Split(s, ";"); len(parts) > 0 {
		s = strings.TrimSpace(parts[0])
	}

	// 3. 匹配
	switch s {
	case "text/plain":
		return ContentTypePlainText
	case "text/html", "application/xhtml+xml":
		return ContentTypeHTML
	case "application/json", "text/javascript":
		return ContentTypeJSON
	case "text/xml", "application/xml":
		return ContentTypeXML
	case "application/x-www-form-urlencoded":
		return ContentTypeForm
	case "text/event-stream":
		return ContentTypeEventStream
	// 2026 建議：multipart 通常帶有 boundary，用 HasPrefix 或等於判斷
	case "multipart/form-data":
		return ContentTypeMultipart
	default:
		return ContentTypeUnknown
	}
}

func (that contentTyper) SetContentType(contentType TypeContent) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), ContentTypeCtxKey, contentType))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// GetRequestContentType is a helper function that returns ContentType based on
// context or request headers.
func (that contentTyper) GetRequestContentType(r *http.Request) TypeContent {
	if contentType, ok := r.Context().Value(ContentTypeCtxKey).(TypeContent); ok {
		return contentType
	}
	return GetContentType(r.Header.Get("Content-Type"))
}

func (that contentTyper) GetAcceptedContentType(r *http.Request) TypeContent {
	if contentType, ok := r.Context().Value(ContentTypeCtxKey).(TypeContent); ok {
		return contentType
	}

	var contentType TypeContent

	// Parse request Accept header.
	fields := strings.Split(r.Header.Get("Accept"), ",")
	if len(fields) > 0 {
		contentType = GetContentType(strings.TrimSpace(fields[0]))
	}

	if contentType == ContentTypeUnknown {
		contentType = ContentTypePlainText
	}
	return contentType
}
