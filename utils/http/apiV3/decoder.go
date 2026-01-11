package apiV3

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin/binding"
)

var Decoder = decoder{}

type decoder struct{}

// Decode 2026 优化版：增加了流保护和更稳健的类型识别
func (d decoder) Decode(r *http.Request, v interface{}) (resp interface{}, err error) {
	if r.Body == nil {
		return nil, errors.New("request body is empty")
	}

	// 2026 安全实践：限制读取大小，防止恶意大报文攻击
	// 默认限制 10MB，可根据项目需求调整
	r.Body = http.MaxBytesReader(nil, r.Body, 10<<20)

	conType := ContentTyper.GetRequestContentType(r)

	switch conType {
	case ContentTypeJSON:
		err = json.NewDecoder(r.Body).Decode(v)
	case ContentTypeXML:
		err = xml.NewDecoder(r.Body).Decode(v)
	case ContentTypeForm, ContentTypeMultipart:
		// 2026 推荐：对于 Form 和 Multipart，统一交给成熟的 binding 组件
		// 它会自动处理 r.ParseForm() 或 r.ParseMultipartForm()
		err = binding.Form.Bind(r, v)
	default:
		// 使用 fmt.Errorf 替代 errors.New(fmt.Sprintf)
		err = fmt.Errorf("apiV3: unsupported content type [%v]", conType)
	}

	return v, err
}

// DecodeJSON 等私有方法建议保留，但移除不必要的 io.Discard
func (d decoder) DecodeJSON(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}
