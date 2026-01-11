package apiV3

import (
	"net/http"
	"strconv"
)

// GetHeaderValue 获取 Header 值
func GetHeaderValue(r *http.Request, key string) string {
	return r.Header.Get(key)
}

// GetUid 获取 UID (int64 是 2026 年 ID 处理的标准类型)
func GetUid(r *http.Request) int64 {
	val := GetHeaderValue(r, "jwt_uid")
	if val == "" {
		return 0
	}
	uid, _ := strconv.ParseInt(val, 10, 64)
	return uid
}

// GetUidV2 泛型优化版
func GetUidV2[T int | int32 | int64](r *http.Request) T {
	uidStr := GetHeaderValue(r, "jwt_uid")
	if uidStr == "" {
		return 0
	}

	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		return 0
	}

	// 2026 推荐写法：简洁的类型转换
	return T(uid)
}

// GetAppId 获取 AppId
func GetAppId(r *http.Request) string {
	return GetHeaderValue(r, "jwt_appid")
}

// GetCid 获取 CID
func GetCid(r *http.Request) int32 {
	cid, _ := strconv.ParseInt(GetHeaderValue(r, "jwt_cid"), 10, 32)
	return int32(cid)
}

// QueryString 获取 Query 参数
func QueryString(r *http.Request, key string) string {
	// 2026 性能提示：r.URL.Query() 每次调用都会重新解析字符串
	// 如果频繁调用，建议在 Context 中缓存该值
	return r.URL.Query().Get(key)
}

// QueryInt32 获取 int32 类型的 Query 参数
func QueryInt32(r *http.Request, key string) int32 {
	v := r.URL.Query().Get(key)
	if v == "" {
		return 0
	}
	val, _ := strconv.ParseInt(v, 10, 32)
	return int32(val)
}

// QueryInt64 补充 2026 常用类型
func QueryInt64(r *http.Request, key string) int64 {
	v := r.URL.Query().Get(key)
	val, _ := strconv.ParseInt(v, 10, 64)
	return val
}
