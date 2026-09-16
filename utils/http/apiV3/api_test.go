package apiV3

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cute-angelia/go-xutils/utils/iAes"
)

func TestApiV3_Success_AES_GCM(t *testing.T) {
	cryptoKey := "1234567890123456" // 16 bytes
	req, _ := http.NewRequest("GET", "/test?crypto=3", nil)
	rr := httptest.NewRecorder()

	app := NewApi(rr, req, WithCryptoKey(cryptoKey))
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
