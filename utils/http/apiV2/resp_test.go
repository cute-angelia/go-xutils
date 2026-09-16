package apiV2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cute-angelia/go-xutils/utils/iAes"
)

func TestApiV2_SuccessEncrypt_AES_GCM(t *testing.T) {
	cryptoKey := "1234567890123456" // 16 bytes
	reqData := map[string]interface{}{
		"userId": 20088,
		"msg":    "apiV2 gcm test",
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

	if int(decryptedData["userId"].(float64)) != 20088 || decryptedData["msg"] != "apiV2 gcm test" {
		t.Fatalf("Decrypted data mismatch: %v", decryptedData)
	}
}

func TestApiV2_SuccessEncryptWithPage_AES_GCM(t *testing.T) {
	cryptoKey := "1234567890123456"
	reqData := []string{"item1", "item2"}
	pager := Pagination{
		Current:  1,
		PageSize: 10,
		Total:    2,
	}

	req, _ := http.NewRequest("GET", "/test?crypto=3", nil)
	rr := httptest.NewRecorder()

	SuccessEncryptWithPage(rr, req, reqData, "ok", pager, cryptoKey)

	var res ResPage
	_ = json.NewDecoder(rr.Body).Decode(&res)

	cipherStr := res.Data.(string)
	randomKey := cipherStr[:16]
	ciphertextBase64 := cipherStr[16:]
	cryptoId := fmt.Sprintf("%s%s", cryptoKey, randomKey)

	decryptedBytes, err := iAes.DecryptGCMFromBase64(ciphertextBase64, []byte(cryptoId))
	if err != nil {
		t.Fatalf("DecryptGCMFromBase64 failed: %v", err)
	}

	var list []string
	_ = json.Unmarshal(decryptedBytes, &list)
	if len(list) != 2 || list[0] != "item1" {
		t.Fatalf("Decrypted list mismatch: %v", list)
	}
}
