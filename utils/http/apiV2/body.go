package apiV2

import (
	"bytes"
	"encoding/json"
	jsoniter "github.com/json-iterator/go"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	ContentTypeWWWForm = "application/x-www-form-urlencoded"
	ContentTypeJson    = "application/json"
)

type body struct {
	r           *http.Request
	isJson      bool
	dataJson    map[string]interface{}
	dataWWWForm url.Values
}

var jsonLib = jsoniter.ConfigCompatibleWithStandardLibrary

func NewBody(r *http.Request) *body {

	contentType := r.Header.Get("Content-Type")

	// 过滤分号
	//application/x-www-form-urlencoded; charset=UTF-8
	if strings.Contains(contentType, ";") {
		contentType = strings.Split(contentType, ";")[0]
	}

	b := body{
		r: r,
	}
	if contentType == ContentTypeWWWForm {
		if r.PostForm == nil {
			r.ParseMultipartForm(32 << 20) // 32MB
		}
		b.dataWWWForm = r.Form
	}
	if contentType == ContentTypeJson || strings.Contains(contentType, ContentTypeJson) {
		b.isJson = true

		buf, _ := io.ReadAll(r.Body)

		d := jsonLib.NewDecoder(bytes.NewReader(buf))
		d.UseNumber()
		d.Decode(&b.dataJson)

		r.Body = io.NopCloser(bytes.NewReader(buf))
	}
	return &b
}

// from post
func (b *body) PostBody() (val string) {
	if b.isJson {
		val1, _ := jsonLib.Marshal(b.dataJson)
		return string(val1)
	} else {
		val2, _ := jsonLib.Marshal(b.dataWWWForm)
		return string(val2)
	}
}

func (b *body) PostString(key string) (val string) {
	if b.isJson {
		if value, ok := b.dataJson[key]; ok {
			if strValue, ok := value.(string); ok {
				val = strValue
			}
		}
	} else {
		val = b.dataWWWForm.Get(key)
	}
	return val
}

// PostSlice 数组，组合可能很多，有map ，strings ，int
func (b *body) PostSlice(key string) (val []interface{}) {
	if b.isJson {
		if _, ok := b.dataJson[key]; !ok {
			return []interface{}{}
		} else {
			val = b.dataJson[key].([]interface{})
		}
		log.Println(val)
	} else {
		val = append(val, b.dataWWWForm.Get(key))
	}
	return val
}

func (b *body) PostBool(key string) (val bool) {
	if b.isJson {
		if value, ok := b.dataJson[key]; ok {
			if strValue, ok := value.(bool); ok {
				val = strValue
			}
		}
	} else {
		val, _ = strconv.ParseBool(b.dataWWWForm.Get(key))
	}
	return val
}

func (b *body) PostInt(key string) (val int) {
	if b.isJson {
		if value, ok := b.dataJson[key]; ok {
			if strValue, err := value.(json.Number).Int64(); err == nil {
				val = int(strValue)
			} else {
				log.Println(err)
			}
		}
	} else {
		ageInt, _ := strconv.Atoi(b.dataWWWForm.Get(key))
		val = ageInt
	}
	return val
}
func (b *body) PostInt32(key string) (val int32) {
	if b.isJson {
		if value, ok := b.dataJson[key]; ok {
			if strValue, err := value.(json.Number).Int64(); err == nil {
				val = int32(strValue)
			} else {
				log.Println(err)
			}
		}
	} else {
		ageInt, _ := strconv.Atoi(b.dataWWWForm.Get(key))
		val = int32(ageInt)
	}
	return val
}
func (b *body) PostInt64(key string) (val int64) {
	if b.isJson {
		if value, ok := b.dataJson[key]; ok {
			if strValue, err := value.(json.Number).Int64(); err == nil {
				val = strValue
			} else {
				log.Println(err)
			}
		}
	} else {
		ageInt, _ := strconv.Atoi(b.dataWWWForm.Get(key))
		val = int64(ageInt)
	}
	return val
}
func (b *body) PostFloat64(key string) (val float64) {
	if b.isJson {
		if value, ok := b.dataJson[key]; ok {
			if strValue, err := value.(json.Number).Float64(); err == nil {
				val = strValue
			}
		}
	} else {
		ageInt, _ := strconv.ParseFloat(b.dataWWWForm.Get(key), 10)
		val = ageInt
	}
	return val
}

// from upload

func (b *body) UploadParseMultipartForm(maxMemory int64) error {
	return b.r.ParseMultipartForm(maxMemory)
}

func (b *body) Upload(key string) string {
	return b.r.FormValue(key)
}

func (b *body) Upload32(key string) int32 {
	p, _ := strconv.Atoi(b.r.FormValue(key))
	return int32(p)
}

func (b *body) PostArray(key string) []string {
	return b.r.PostForm[key]
}
