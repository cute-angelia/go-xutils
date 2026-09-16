package iAes

import (
	"bytes"
	"testing"
)

func TestAES_CBC(t *testing.T) {
	key := []byte("12345678901234567890123456789012") // 32 bytes AES-256
	plaintext := []byte("hello world, testing AES CBC encryption and decryption!")

	// Base64
	encryptedBase64, err := EncryptCBCToBase64(plaintext, key)
	if err != nil {
		t.Fatalf("EncryptCBCToBase64 failed: %v", err)
	}

	decrypted, err := DecryptCBCFromBase64(encryptedBase64, key)
	if err != nil {
		t.Fatalf("DecryptCBCFromBase64 failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("Decrypted mismatch: got %s, want %s", string(decrypted), string(plaintext))
	}
}

func TestAES_GCM(t *testing.T) {
	key := []byte("12345678901234567890123456789012") // 32 bytes AES-256
	plaintext := []byte("{\"code\":0,\"msg\":\"success\",\"data\":{\"userId\":12345,\"name\":\"vanilla\"}}")

	// 1. Raw bytes GCM
	encrypted, err := EncryptGCM(plaintext, key)
	if err != nil {
		t.Fatalf("EncryptGCM failed: %v", err)
	}

	decrypted, err := DecryptGCM(encrypted, key)
	if err != nil {
		t.Fatalf("DecryptGCM failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("Decrypted mismatch: got %s, want %s", string(decrypted), string(plaintext))
	}

	// 2. Base64 GCM
	encryptedBase64, err := EncryptGCMToBase64(plaintext, key)
	if err != nil {
		t.Fatalf("EncryptGCMToBase64 failed: %v", err)
	}

	decryptedBase64, err := DecryptGCMFromBase64(encryptedBase64, key)
	if err != nil {
		t.Fatalf("DecryptGCMFromBase64 failed: %v", err)
	}

	if !bytes.Equal(plaintext, decryptedBase64) {
		t.Fatalf("Decrypted base64 mismatch: got %s, want %s", string(decryptedBase64), string(plaintext))
	}

	// 3. Test AEAD authentication (tampering detection)
	tampered := make([]byte, len(encrypted))
	copy(tampered, encrypted)
	tampered[len(tampered)-1] ^= 0x01 // flip a bit in auth tag

	_, err = DecryptGCM(tampered, key)
	if err == nil {
		t.Fatalf("Expected decryption to fail on tampered ciphertext, but it succeeded!")
	}
}
