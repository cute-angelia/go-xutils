package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"github.com/cute-angelia/go-xutils/utils/generator/base"
	"io"
	"log"
	"testing"
)

var key = []byte("secret key 123xx")
var msg = []byte("good你好")

// go test -v -run TestEncrypt
func TestEncrypt(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	iaes, _ := NewAesPackage([]byte(""))
	cipherstring := iaes.EncryptCFB(msg).ToStringBase64()
	log.Println("EncryptCFB", cipherstring)
	decodestring, _ := base.Base64Decode(cipherstring)
	log.Println("DecryptCFB", iaes.DecryptCFB(decodestring).ToString())

	// cbc
	cipherstring2 := iaes.EncryptCBC(msg, PaddingPkcs7).ToStringBase64()
	log.Println("EncryptCBC", cipherstring2)
	cbcstr, _ := base.Base64Decode(cipherstring2)
	log.Println("DecryptCBC", iaes.DecryptCBC(cbcstr, iaes.CurrentCipher, PaddingPkcs7).ToString())
}

func TestJsDecrypt(t *testing.T) {
	ikey := "5a673bd785831e2e5a673bd785831e2e"
	//crypted := "ea17db2f3930c4f94b425141832f7f68"

	// js hex 解码
	//cipherText, err := hex.DecodeString(crypted)
	//if err != nil {
	//	log.Println(err)
	//}

	iaes, _ := NewAesPackage([]byte(ikey))
	//dd := iaes.DecryptCBC(cipherText, PaddingPkcs7).ToString()
	//log.Println("DecryptCBC", dd)

	// 目前 iv 算法取 secret 前16位
	log.Println("EncryptCBC", iaes.EncryptCBC([]byte("hello world --> replay"), PaddingPkcs7).ToStringHex())
}

// 使用AES-GCM模式,处理密钥、认证、加密一次完成
// 注意，nodejs 需要 tag ，需要返回tag
func encryptGcm(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func encryptGCMWithTag(plaintext []byte, key []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	gcm, err := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	// 取出tag
	tag := ciphertext[len(ciphertext)-gcm.Overhead():]
	return ciphertext, tag, err
}
