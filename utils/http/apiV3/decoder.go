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
	// 1. 如果存在 URL Query 参数，先解码 Query（兼容签名参数与 Query 传参）
	if len(r.URL.Query()) > 0 {
		_ = queryDecoder.Decode(v, r.URL.Query())
	}

	// GET 请求直接返回
	if r.Method == http.MethodGet {
		return v, nil
	}

	// Body 为空时容错（例如无 body 的 POST），直接返回已解析的 query
	if r.Body == nil {
		return v, nil
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
