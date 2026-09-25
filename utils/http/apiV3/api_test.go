package apiV3

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cute-angelia/go-xutils/utils/iAes"
)

func TestApiV3_Success_AES_GCM(t *testing.T) {
	cryptoKey := "1234567890123456" // 16 bytes
	req, _ := http.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// 明确指定 AES-GCM，而不是依赖 URL query 参数（服务端配置了 key 后会忽略客户端 query）
	app := NewApi(rr, req, WithCryptoKey(cryptoKey), WithCryptoType(CryptoTypeAESGCM))
	app.SetData(map[string]interface{}{
		"name": "gcm_test",
		"id":   888,
	})
	app.SetMsg("gcm ok")
	app.Success()

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", rr.Code)
	}

	var res Res
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	cipherStr, ok := res.Data.(string)
	if !ok || len(cipherStr) <= 16 {
		t.Fatalf("Expected cipher string, got %v", res.Data)
	}

	randomKey := cipherStr[:16]
	ciphertextBase64 := cipherStr[16:]
	cryptoId := fmt.Sprintf("%s%s", cryptoKey, randomKey)

	decryptedBytes, err := iAes.DecryptGCMFromBase64(ciphertextBase64, []byte(cryptoId))
	if err != nil {
		t.Fatalf("DecryptGCMFromBase64 failed: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(decryptedBytes, &data); err != nil {
		t.Fatalf("Unmarshal decrypted failed: %v", err)
	}

	if data["name"] != "gcm_test" || int(data["id"].(float64)) != 888 {
		t.Fatalf("Mismatch decrypted data: %v", data)
	}
}

func TestApiV3_Success_AES_GCM_ByOption(t *testing.T) {
	cryptoKey := "1234567890123456"
	req, _ := http.NewRequest("GET", "/test?crypto=true", nil)
	rr := httptest.NewRecorder()

	app := NewApi(rr, req, WithCryptoKey(cryptoKey), WithCryptoType(CryptoTypeAESGCM))
	app.SetData("secret message")
	app.Success()

	var res Res
	_ = json.NewDecoder(rr.Body).Decode(&res)

	cipherStr := res.Data.(string)
	randomKey := cipherStr[:16]
	ciphertextBase64 := cipherStr[16:]
	cryptoId := fmt.Sprintf("%s%s", cryptoKey, randomKey)

	decryptedBytes, err := iAes.DecryptGCMFromBase64(ciphertextBase64, []byte(cryptoId))
	if err != nil {
		t.Fatalf("DecryptGCMFromBase64 failed: %v", err)
	}

	var text string
	_ = json.Unmarshal(decryptedBytes, &text)
	if text != "secret message" {
		t.Fatalf("Mismatch: %s", text)
	}
}

// ---- Validation 测试 ----

type testReq struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// TestValidation_GET_ReadsQuery GET 请求 Validation 应从 URL Query 读参数
func TestValidation_GET_ReadsQuery(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test?name=alice&age=30", nil)
	rr := httptest.NewRecorder()

	render := NewApi(rr, req)
	v := testReq{}
	if err := render.Validation(&v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Name != "alice" || v.Age != 30 {
		t.Fatalf("expected name=alice age=30, got %+v", v)
	}
}

// TestValidation_POST_OnlyReadsBody POST 请求 Validation 只读 Body，URL Query 不应被并入
func TestValidation_POST_OnlyReadsBody(t *testing.T) {
	body := `{"name":"bob","age":25}`
	// URL 上故意带 name=hacker，不应被写入结构体
	req, _ := http.NewRequest("POST", "/test?name=hacker", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	render := NewApi(rr, req)
	v := testReq{}
	if err := render.Validation(&v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Name != "bob" {
		t.Fatalf("expected name=bob from body, got %q (URL query must not bleed in)", v.Name)
	}
}

// TestValidationFromQuery_POST_ReadsQuery POST 时手动调 ValidationFromQuery 应读取 URL Query
func TestValidationFromQuery_POST_ReadsQuery(t *testing.T) {
	body := `{"name":"ignored"}`
	req, _ := http.NewRequest("POST", "/test?name=query_name&age=99", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	render := NewApi(rr, req)
	v := testReq{}
	if err := render.ValidationFromQuery(&v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Name != "query_name" || v.Age != 99 {
		t.Fatalf("expected query_name/99 from URL query, got %+v", v)
	}
}
