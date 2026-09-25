package apiV3

import (
	"encoding/xml"
	"io"
	"log"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gorilla/schema"
)

var Decoder = decoder{}
var queryDecoder = schema.NewDecoder()

func init() {
	queryDecoder.IgnoreUnknownKeys(true)
	// 告訴 schema 插件，去讀取結構體上的 "form" tag
	queryDecoder.SetAliasTag("json") // schema, form
}

type decoder struct{}

func (d decoder) Decode(r *http.Request, v interface{}) (resp interface{}, err error) {
	// GET 请求走 DecodeQuery
	if r.Method == http.MethodGet {
		return d.DecodeQuery(r, v)
	}

	// Body 为空时容错（例如无 body 的 POST），直接返回
	if r.Body == nil {
		return v, nil
	}

	// 安全限制：用 io.LimitReader 代替 http.MaxBytesReader，
	// 避免因第一个参数为 nil 的 ResponseWriter 在超限时 panic
	limitedBody := io.LimitReader(r.Body, 10<<20)

	conType := ContentTyper.GetRequestContentType(r)

	switch {
	case conType == ContentTypeJSON:
		data, err := io.ReadAll(limitedBody)
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			err = sonic.ConfigDefault.Unmarshal(data, v)
		}

	case conType == ContentTypeXML:
		err = xml.NewDecoder(r.Body).Decode(v)

	case conType == ContentTypeForm:
		if len(r.PostForm) > 0 {
			err = queryDecoder.Decode(v, r.PostForm)
		} else {
			if err := r.ParseForm(); err != nil {
				log.Println(err)
			}
			err = queryDecoder.Decode(v, r.PostForm)
		}

	case conType == ContentTypeMultipart:
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return nil, err
		}
		err = queryDecoder.Decode(v, r.MultipartForm.Value)

	default:
		// 未指定 Content-Type 时尝试当 JSON 解析，无内容则跳过
		data, errRead := io.ReadAll(r.Body)
		if errRead == nil && len(data) > 0 {
			_ = sonic.ConfigDefault.Unmarshal(data, v)
		}
	}
	return v, err
}

// DecodeQuery 只从 URL Query 参数解析，不读取 Body
func (d decoder) DecodeQuery(r *http.Request, v interface{}) (resp interface{}, err error) {
	if len(r.URL.Query()) > 0 {
		err = queryDecoder.Decode(v, r.URL.Query())
	}
	return v, err
}
