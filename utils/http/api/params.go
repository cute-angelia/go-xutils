package api

import (
	"github.com/cute-angelia/go-xutils/utils/generator/hash"
	"github.com/go-chi/chi"
	jsoniter "github.com/json-iterator/go"
	"net/http"
	"strconv"
)

// Query will get a query parameter by key.
func QueryString(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

// QueryInt will get a query parameter by key and convert it to an int or return an error.
// user?user_id=
func QueryStringInt32(r *http.Request, key string) int32 {
	v := r.URL.Query().Get(key)
	val, _ := strconv.Atoi(v)
	return int32(val)
}

// user/{user_id}
func Param(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

func ParamInt32(r *http.Request, name string) int32 {
	p, _ := strconv.Atoi(chi.URLParam(r, name))
	return int32(p)
}

func UploadParseMultipartForm(r *http.Request, maxMemory int64) error {
	return r.ParseMultipartForm(maxMemory)
}

func Upload(r *http.Request, name string) string {
	return r.FormValue(name)
}

func Upload32(r *http.Request, name string) int32 {
	p, _ := strconv.Atoi(r.FormValue(name))
	return int32(p)
}

func Post(r *http.Request, name string) string {
	//if len(requireds) > 0 {
	//	if vs := r.PostForm[name]; len(vs) > 0 {
	//		return vs[0], nil
	//	} else {
	//		return "", fmt.Errorf("缺少必传参数:" + name)
	//	}
	//}
	return r.PostFormValue(name)
}

func PostArray(r *http.Request, name string) []string {
	return r.PostForm[name]
}

func PostInt32(r *http.Request, name string) int32 {
	p, _ := strconv.Atoi(r.PostFormValue(name))
	return int32(p)
}

func PostInt(r *http.Request, name string) int {
	p, _ := strconv.Atoi(r.PostFormValue(name))
	return p
}

func PostInt64(r *http.Request, name string) int64 {
	p, _ := strconv.Atoi(r.PostFormValue(name))
	return int64(p)
}

func PostFloat64(r *http.Request, name string) float64 {
	p, _ := strconv.ParseFloat(r.PostFormValue(name), 64)
	return p
}

func GetUserUid(r *http.Request) int32 {
	uid, _ := strconv.Atoi(r.Header.Get("jwt_uid"))
	return int32(uid)
}

func GetUserUidInt64(r *http.Request) int64 {
	uid, _ := strconv.Atoi(r.Header.Get("jwt_uid"))
	return int64(uid)
}

func GetAppId(r *http.Request) string {
	appid := r.Header.Get("jwt_appid")
	return appid
}

func GetCid(r *http.Request) int32 {
	jwt_cid, _ := strconv.Atoi(r.Header.Get("jwt_cid"))
	return int32(jwt_cid)
}

// GetJwtHeader 获取 jwt 里面数据
func GetJwtHeader(r *http.Request, key string) interface{} {
	return r.Header.Get(key)
}

// GenerateCacheKey 缓存key
func GenerateCacheKey(params interface{}, filtes []string, prefix string) string {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	bparms, _ := json.Marshal(params)
	m := make(map[string]interface{})
	json.Unmarshal(bparms, &m)

	for _, filte := range filtes {
		if _, ok := m[filte]; ok {
			delete(m, filte)
		}
	}
	cacheKey, _ := json.Marshal(m)
	return prefix + "_" + hash.Hash(hash.AlgoMD5, string(cacheKey))
}
