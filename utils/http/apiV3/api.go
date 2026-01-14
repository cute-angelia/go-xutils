package apiV3

import (
	"encoding/json"
	"errors"
	"github.com/cute-angelia/go-xutils/syntax/irandom"
	"github.com/cute-angelia/go-xutils/utils/iAes"
	"github.com/cute-angelia/go-xutils/utils/iXor"
	"github.com/go-ozzo/ozzo-validation/v4"
	"log"
	"net/http"
	"strconv"
	"time"
)

type api struct {
	w http.ResponseWriter
	r *http.Request

	isHasPage bool // 是否分页
	pager     Pagination

	cryptoType CryptoType // 加密方式：默认2
	cryptoKey  string     // 是否加密：不为空为加密

	isLogOn bool // 打印日志

	reqStruct  any // 请求结构体
	respStruct Res // 返回结构体
}

// Res 标准JSON输出格式
type Res struct {
	// Code 响应的业务错误码。0表示业务执行成功，非0表示业务执行失败。
	Code int32 `json:"code"`
	// Msg 响应的参考消息。前端可使用msg来做提示
	Msg string `json:"msg"`
	// Data 响应的具体数据
	Data interface{} `json:"data,omitempty"`

	Pagination *Pagination `json:"pagination,omitempty"`

	Ext *Ext `json:"ext,omitempty"`
}

type Ext struct {
	ShowTips bool `json:"showTips"` // 弹消息提示
}

// Pagination 分页结构体
type Pagination struct {
	//  当前页
	PageNo int64 `json:"pageNo"`
	// PageSize 每页记录数
	PageSize int64 `json:"pageSize"`
	// PageTotal 总页数
	PageTotal int64 `json:"pageTotal"`
	// 总条数
	Count int64 `json:"count"`
}

// CalcTotal 计算总页数
func (p Pagination) CalcTotal(count, pageSize int64) int64 {
	if pageSize <= 0 {
		return 0
	}
	return (count + pageSize - 1) / pageSize
}

func NewPagination(count, pageNo, pageSize int64) Pagination {
	paginationor := Pagination{PageNo: pageNo, PageSize: pageSize, Count: count}
	paginationor.PageTotal = paginationor.CalcTotal(count, pageSize)
	return paginationor
}

// Pagination 分页结构体 end
func NewApi(w http.ResponseWriter, r *http.Request, opts ...Option) *api {
	a := &api{
		w:          w,
		r:          r,
		isLogOn:    true,                              // 默認值
		cryptoType: CryptoTypeXOR,                     // 默認值
		cryptoKey:  CryptoEr.GetRequestContentType(r), // 默認值
	}
	// 應用所有傳入的選項
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Decode request
func (that *api) Decode(v interface{}) error {
	body, err := Decoder.Decode(that.r, v)
	that.reqStruct = body
	return err
}

// Validation 接收 v 以及可選的額外規則
/*
req := RotateReq{}

方式一：
err := render.Validation(&req,
    validation.Field(&req.MinioFullUrl, validation.Required, validation.Match(regexp.MustCompile(`^https?://`))),
    validation.Field(&req.Position, validation.In(90, 180, 270)),
)
if err != nil {
    render.Error(err)
    return
}


方式二：
// 實現接口，這樣 ValidateStruct 會自動調用這裡
func (r RotateReq) Validate() error {
    return validation.ValidateStruct(&r,
        validation.Field(&r.MinioFullUrl, validation.Required),
        validation.Field(&r.Position, validation.In(90, 180, 270)),
    )
}

*/
func (that *api) Validation(v interface{}, fields ...*validation.FieldRules) error {
	// 1. 解析請求數據 (JSON/Form/Query)
	if err := that.Decode(v); err != nil {
		return err
	}

	// 2. 執行結構體基礎校驗 (基於結構體內的 Internal Validate 方法)
	// 如果結構體實現了 validation.Validatable 接口，會自動執行
	if err := validation.ValidateStruct(v, fields...); err != nil {
		return err
	}

	return nil
}

// ValidMustLogin 檢查登入狀態，若未登入則輸出錯誤並返回 false
func (that *api) ValidMustLogin() bool {
	uid := that.GetUid()
	if uid <= 0 {
		that.ErrorCodeMsg(-401, "請先登入")
		return false
	}
	return true
}

// GetUid 從 Header 中獲取 JWT 解析後的 UID
func (that *api) GetUid() int32 {
	val := that.r.Header.Get("jwt_uid")
	if val == "" {
		return 0
	}
	uid, err := strconv.Atoi(val)
	if err != nil {
		// 2026 實踐：記錄日誌但不中斷流程，返回 0 代表無效用戶
		log.Printf("apiV3: invalid jwt_uid header: %v", val)
		return 0
	}
	return int32(uid)
}

func (that *api) SetData(data interface{}) *api {
	that.respStruct.Data = data
	return that
}

func (that *api) SetMsg(msg string) *api {
	that.respStruct.Msg = msg
	return that
}

func (that *api) SetPage(pager *Pagination) *api {
	that.respStruct.Pagination = pager
	return that
}

func (that *api) SetExt(ext *Ext) *api {
	that.respStruct.Ext = ext
	return that
}

// Success 成功返回
func (that *api) Success() {
	that.respStruct.Code = 0
	// 日志
	that.logr("[success]")

	// 加密
	that.cryptoData()

	// json
	that.w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(that.w).Encode(that.respStruct); err != nil {
		// 2026 实践：编码失败属于服务器内部错误，记录日志但不一定发送给客户端
		log.Printf("JSON Encode Error: %v", err)
		http.Error(that.w, "Internal Server Error", 500)
		return
	}
}

func (that *api) ErrorCodeMsg(code int32, msg string) {
	err := NewApiError(code, msg)
	that.Error(err)
}

func (that *api) Error(err error) {
	that.respStruct.Code = -1

	if err != nil {
		var e *ApiError
		if errors.As(err, &e) {
			// 可以访问e.Code和e.Message
			that.respStruct.Code = e.Code
		}
		that.respStruct.Msg = err.Error()
	}

	that.logr("[error]")

	// 加密
	that.cryptoData()

	// json
	that.w.Header().Set("Content-Type", "application/json")

	if err = json.NewEncoder(that.w).Encode(that.respStruct); err != nil {
		http.Error(that.w, err.Error(), 500)
		return
	}
}

func (that *api) cryptoData() {
	crypto := that.r.URL.Query().Get("crypto")
	if len(crypto) > 0 {
		var randomKey = irandom.RandString(16, irandom.LetterAll)
		cryptoId := that.cryptoKey + randomKey
		datam, _ := json.Marshal(that.respStruct.Data)

		// Crypto 加密 Key：使用AES-GCM模式,处理密钥、认证、加密一次完成
		if that.cryptoType == 1 {
			encryptData, _ := iAes.EncryptCBCToBase64(datam, []byte(cryptoId))
			that.respStruct.Data = randomKey + encryptData
		}
		// xor
		if that.cryptoType == 2 {
			encryptData := iXor.XorEncrypt(datam, cryptoId)
			that.respStruct.Data = randomKey + encryptData
		}
	}
}

func (that *api) logr(tag string) {
	defer func() { recover() }()

	// 为了不破坏 respStruct 的 Data 类型，这里局部序列化
	dataReq, _ := json.Marshal(that.reqStruct)
	dataResp, _ := json.Marshal(that.respStruct)

	uid := that.r.Header.Get("jwt_uid")
	appStartTime := that.r.Header.Get("jwt_app_start_time")

	costMsg := ""
	if len(appStartTime) > 0 {
		// 2026 修正：处理毫秒级时间戳
		if un, err := strconv.ParseInt(appStartTime, 10, 64); err == nil {
			t2 := time.UnixMilli(un) // 假设前端传的是毫秒
			if un < 2000000000 {
				t2 = time.Unix(un, 0)
			} // 兼容秒
			cost := time.Since(t2)
			costMsg = "| Cost: " + cost.String()
		}
	}

	log.Println("------------------------------------------------------------------------------")
	log.Printf("%s 用户: %s %s", tag, uid, costMsg)
	log.Printf("地址: %s, 数据: %s", that.r.URL.Path, dataReq)

	if that.isLogOn {
		log.Printf("响应: %s", dataResp)
	} else {
		log.Printf("响应: %s", "關閉列印")
	}
	log.Println("------------------------------------------------------------------------------")
}
