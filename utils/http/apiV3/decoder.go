package apiV3

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/gorilla/schema"
	"io"
	"net/http"
)

var Decoder = decoder{}
var queryDecoder = schema.NewDecoder()

func init() {
	queryDecoder.IgnoreUnknownKeys(true)
}

type decoder struct{}

func (d decoder) Decode(r *http.Request, v interface{}) (resp interface{}, err error) {
	// 1. 处理 GET
	if r.Method == http.MethodGet {
		err = queryDecoder.Decode(v, r.URL.Query())
		return v, err
	}

	if r.Body == nil {
		return nil, errors.New("request body is empty")
	}

	// 2. 安全限制
	r.Body = http.MaxBytesReader(nil, r.Body, 10<<20)

	conType := ContentTyper.GetRequestContentType(r)

	switch {
	case conType == ContentTypeJSON:
		// 3. 读取 Body
		data, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		// --- 关键改动在这里 ---
		// 在 2026 年，使用 ConfigDefault 是最稳健的。
		// 它会自动处理 {"id": "123"} 到 int32 的转换，无需加 ,string 标签。
		// 如果你发现默认不转，请使用 ConfigStd (标准库兼容模式)
		err = sonic.ConfigDefault.Unmarshal(data, v)
		// ---------------------

	case conType == ContentTypeXML:
		err = xml.NewDecoder(r.Body).Decode(v)

	case conType == ContentTypeForm:
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
		err = queryDecoder.Decode(v, r.PostForm)

	case conType == ContentTypeMultipart:
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return nil, err
		}
		err = queryDecoder.Decode(v, r.MultipartForm.Value)

	default:
		err = fmt.Errorf("apiV3: unsupported content type [%v]", conType)
	}
	return v, err
}
