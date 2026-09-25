package apiV3

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/cute-angelia/go-xutils/syntax/irandom"
	"github.com/cute-angelia/go-xutils/utils/iAes"
	"github.com/cute-angelia/go-xutils/utils/iXor"
	"github.com/go-ozzo/ozzo-validation/v4"
)

type Api struct {
	w http.ResponseWriter
	r *http.Request

	cryptoType CryptoType // 加密方式：默认2
	cryptoKey  string     // 是否加密：不为空为加密

	isLogOn bool // 打印日志

	reqStruct  any // 请求结构体
	respStruct Res // 返回结构体
}

type api = Api

// Res 标准JSON输出格式
type Res struct {
	// Code 响应的业务错误码。0表示业务执行成功，非0表示业务执行失败。
	Code int32 `json:"code"`
	// Msg 响应的参考消息。前端可使用msg来做提示
	Msg string `json:"msg"`
	// Data 响应的具体数据
	Data interface{} `json:"data"`

	Pagination *Pagination `json:"pagination,omitempty"`

	Ext *Ext `json:"ext,omitempty"`
}

type Ext struct {
	ShowTips bool `json:"showTips"` // 弹消息提示
}

// Pagination 分页结构体
type Pagination struct {
	//  当前页
	Page int64 `json:"page"`
	// PageSize 每页记录数
	PageSize int64 `json:"page_size"`
	// PageTotal 总页数
	PageTotal int64 `json:"page_total"`
	// 总条数
	Total int64 `json:"total"`
}

// CalcTotal 计算总页数
func (p Pagination) CalcTotal(count, pageSize int64) int64 {
	if pageSize <= 0 {
		return 0
	}
	return (count + pageSize - 1) / pageSize
}

func NewPagination(count, Page, pageSize int64) Pagination {
	paginationor := Pagination{Page: Page, PageSize: pageSize, Total: count}
	paginationor.PageTotal = paginationor.CalcTotal(count, pageSize)
	return paginationor
}

// Pagination 分页结构体 end
func NewApi(w http.ResponseWriter, r *http.Request, opts ...Option) *Api {
	a := &Api{
		w:          w,
		r:          r,
		isLogOn:    true,                              // 默認值
		cryptoType: CryptoTypeXOR,                     // 默認值
		cryptoKey:  CryptoEr.GetCryptoKey(r),          // 默認值
	}
	// 應用所有傳入的選項
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// SetReq 手动设置请求结构体或数据，用于日志记录
func (that *Api) SetReq(req any) *Api {
	that.reqStruct = req
	return that
}

// Decode 解析请求数据到 v。
// GET 请求从 URL Query 读取；其他方法只读 Body，与 Validation 行为一致。
func (that *api) Decode(v interface{}) error {
	var body interface{}
	var err error
	if that.r.Method == http.MethodGet {
		body, err = Decoder.DecodeQuery(that.r, v)
	} else {
		body, err = Decoder.Decode(that.r, v)
	}
	that.reqStruct = body
	return err
}

// Validation 解析请求并校验字段。
// GET 请求自动从 URL Query 读取；其他方法（POST/PUT 等）只从 Body 读取，URL Query 不会并入。
// 如需在非 GET 请求中读取 URL Query 参数，请使用 ValidationFromQuery。
func (that *api) Validation(v interface{}, fields ...*validation.FieldRules) error {
	var err error
	if that.r.Method == "GET" {
		_, err = Decoder.DecodeQuery(that.r, v)
	} else {
		_, err = Decoder.Decode(that.r, v)
	}
	that.reqStruct = v
	if err != nil {
		return err
	}

	if err = validation.ValidateStruct(v, fields...); err != nil {
		return err
	}

	return nil
}

// ValidationFromQuery 强制从 URL Query 参数解析并校验，不读取 Body。
// 适用于需要在 POST 等请求中单独读取 URL Query 参数的场景。
func (that *api) ValidationFromQuery(v interface{}, fields ...*validation.FieldRules) error {
	_, err := Decoder.DecodeQuery(that.r, v)
	that.reqStruct = v
	if err != nil {
		return err
	}

	if err = validation.ValidateStruct(v, fields...); err != nil {
		return err
	}

	return nil
}

// ValidMustLogin 檢查登入狀態，若未登入則輸出錯誤並返回 false
func (that *api) ValidMustLogin() bool {
	if that.GetUid() <= 0 {
		that.ErrorCodeMsg(-401, "请先登入")
		return false
	}
	return true
}

// GetUid 從 Header 中獲取 JWT 解析後的 UID，返回 int64
func (that *api) GetUid() int64 {
	val := that.r.Header.Get("jwt_uid")
	if val == "" {
		return 0
	}
	uid, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		log.Printf("apiV3: invalid jwt_uid header: %v", val)
		return 0
	}
	return uid
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
	that.logr("[success]")
	that.cryptoData()
	that.writeJSON()
}

func (that *api) ErrorCodeMsg(code int32, msg string) {
	that.Error(NewApiError(code, msg))
}

func (that *api) Error(err error) {
	that.respStruct.Code = -1

	if err != nil {
		var e *ApiError
		if errors.As(err, &e) {
			that.respStruct.Code = e.Code
		}
		that.respStruct.Msg = err.Error()
	}

	that.logr("[error]")
	that.cryptoData()
	that.writeJSON()
}

// writeJSON 先 Marshal 到内存，成功后一次写入，避免 Header 已发送后再写错误码的双写问题
func (that *api) writeJSON() {
	buf, err := json.Marshal(that.respStruct)
	if err != nil {
		log.Printf("apiV3 writeJSON marshal error: %v", err)
		http.Error(that.w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	that.w.Header().Set("Content-Type", "application/json")
	_, _ = that.w.Write(buf)
}

func (that *api) cryptoData() {
	// 安全策略：服务端已配置 cryptoKey + 非零 cryptoType 时，
	// 强制加密，完全忽略客户端传入的 crypto 参数，防止：
	//   1. crypto=0  绕过加密
	//   2. crypto=1/2 降级为弱加密算法
	if len(that.cryptoKey) > 0 && that.cryptoType > CryptoTypeNone {
		that.doEncrypt()
		return
	}

	// 服务端未配置密钥时，才信任客户端 crypto 参数（降级兼容旧逻辑）
	crypto := that.r.URL.Query().Get("crypto")
	if cType, err := strconv.Atoi(crypto); err == nil && cType > 0 {
		that.cryptoType = CryptoType(cType)
		that.doEncrypt()
	}
}

// doEncrypt 执行实际加密，写入 that.respStruct.Data
func (that *api) doEncrypt() {
	var randomKey = irandom.RandString(16, irandom.LetterAll)
	cryptoId := that.cryptoKey + randomKey
	datam, _ := json.Marshal(that.respStruct.Data)

	switch that.cryptoType {
	case CryptoTypeAES: // 1: AES-CBC
		encryptData, err := iAes.EncryptCBCToBase64(datam, []byte(cryptoId))
		if err != nil {
			log.Println("apiV3 doEncrypt AES-CBC error:", err)
			return
		}
		that.respStruct.Data = randomKey + encryptData

	case CryptoTypeXOR: // 2: XOR
		encryptData := iXor.XorEncrypt(datam, cryptoId)
		that.respStruct.Data = randomKey + encryptData

	case CryptoTypeAESGCM: // 3: AES-GCM (AEAD)
		encryptData, err := iAes.EncryptGCMToBase64(datam, []byte(cryptoId))
		if err != nil {
			log.Println("apiV3 doEncrypt AES-GCM error:", err)
			return
		}
		that.respStruct.Data = randomKey + encryptData
	}
}

func (that *api) logr(tag string) {
	defer func() { recover() }()

	// 为了不破坏 respStruct 的 Data 类型，这里局部序列化
	reqData := that.reqStruct
	if reqData == nil {
		if len(that.r.URL.Query()) > 0 {
			reqData = that.r.URL.Query()
		} else {
			reqData = map[string]interface{}{}
		}
	}
	dataReq, _ := json.Marshal(reqData)
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
		log.Printf("响应: %s", "关闭打印")
	}
	log.Println("------------------------------------------------------------------------------")
}
