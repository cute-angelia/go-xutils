package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cute-angelia/go-xutils/utils/iAes"
)

func TestSuccessEncrypt_AES_GCM(t *testing.T) {
	cryptoKey := "1234567890123456" // 16 bytes
	reqData := map[string]interface{}{
		"userId": 10086,
		"title":  "test message",
	}

	req, _ := http.NewRequest("GET", "/test?crypto=3", nil)
	rr := httptest.NewRecorder()

	SuccessEncrypt(rr, req, reqData, "ok", cryptoKey)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rr.Code)
	}

	var res Res
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatalf("Failed to decode response json: %v", err)
	}

	if res.Code != 0 || res.Msg != "ok" {
		t.Fatalf("Unexpected response header: code=%d, msg=%s", res.Code, res.Msg)
	}

	cipherStr, ok := res.Data.(string)
	if !ok || len(cipherStr) <= 16 {
		t.Fatalf("Expected encrypted data string with length > 16, got %v", res.Data)
	}

	// 验证解密
	randomKey := cipherStr[:16]
	ciphertextBase64 := cipherStr[16:]
	cryptoId := fmt.Sprintf("%s%s", cryptoKey, randomKey)

	decryptedBytes, err := iAes.DecryptGCMFromBase64(ciphertextBase64, []byte(cryptoId))
	if err != nil {
		t.Fatalf("DecryptGCMFromBase64 failed: %v", err)
	}

	var decryptedData map[string]interface{}
	if err := json.Unmarshal(decryptedBytes, &decryptedData); err != nil {
		t.Fatalf("Unmarshal decrypted bytes failed: %v", err)
	}

	if int(decryptedData["userId"].(float64)) != 10086 || decryptedData["title"] != "test message" {
		t.Fatalf("Decrypted data mismatch: %v", decryptedData)
	}
}

func TestSuccessEncrypt_AES_CBC(t *testing.T) {
	cryptoKey := "1234567890123456" // 16 bytes
	reqData := map[string]interface{}{
		"userId": 10086,
	}

	req, _ := http.NewRequest("GET", "/test?crypto=1", nil)
	rr := httptest.NewRecorder()

	SuccessEncrypt(rr, req, reqData, "ok", cryptoKey)

	var res Res
	_ = json.NewDecoder(rr.Body).Decode(&res)

	cipherStr := res.Data.(string)
	randomKey := cipherStr[:16]
	ciphertextBase64 := cipherStr[16:]
	cryptoId := fmt.Sprintf("%s%s", cryptoKey, randomKey)

	decryptedBytes, err := iAes.DecryptCBCFromBase64(ciphertextBase64, []byte(cryptoId))
	if err != nil {
		t.Fatalf("DecryptCBCFromBase64 failed: %v", err)
	}

	var decryptedData map[string]interface{}
	_ = json.Unmarshal(decryptedBytes, &decryptedData)

	if int(decryptedData["userId"].(float64)) != 10086 {
		t.Fatalf("Decrypted data mismatch: %v", decryptedData)
	}
}
