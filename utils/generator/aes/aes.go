package aes

/*
 [wiki](https://zh.wikipedia.org/wiki/%E5%88%86%E7%BB%84%E5%AF%86%E7%A0%81%E5%B7%A5%E4%BD%9C%E6%A8%A1%E5%BC%8F)
（1）AES有5种加密模式，分别是：
 （a）电子密码本（Electronic Codebook Book，ECB）； 照块密码的块大小被分为数个块，并对每个块进行独立加密, 不能很好的隐藏数据模式
 （b）密码块链接（Cipher Block Chaining ，CBC），如果明文长度不是分组长度16字节的整数倍需要进行填充； CBC是最为常用的工作模式 加密串行，解密并行
 （c）计数器模式（Counter，CTR）；
 （d）密文反馈（Cipher FeedBack，CFB）； 模式类似于反向 CBC
 （e）输出反馈（Output FeedBack，OFB）。 然后将其与明文块进行异或，得到密文
（2）AES是对称分组加密算法，每组长度为128bits，即16字节。
（3）AES秘钥的长度只能是16、24或32字节，分别对应三种AES，即AES-128, AES-192和AES-256，三者的区别是加密的轮数不同；

数据填充：
ECB，CBC：块密码只能对确定长度的数据块进行处理，而消息的长度通常是可变的。因此部分模式（即ECB和CBC）需要最后一块在加密前进行填充。
CFB，OFB和CTR：不需要对长度不为密码块大小整数倍的消息进行特别的处理。因为这些模式是通过对块密码的输出与明文进行异或工作的。

PKCS Pkcs7    #5/7 padding strategy.
ANSI AnsiX923 X.923 padding strategy.
ISO Iso10126 padding strategy.
ISO/IEC Iso97971 9797-1 Padding Method 2.
ZeroPadding Zero padding strategy.
NoPadding: Padding;
*/

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"github.com/cute-angelia/go-xutils/utils/generator/base"
	"io"
	"log"
)

type aesPackage struct {
	Secret        []byte
	Block         cipher.Block
	CurrentCipher []byte
}

func NewAesPackage(secretKey []byte) (*aesPackage, error) {
	if len(secretKey) == 0 {
		secretKey = []byte("passphrasewhichneedstobe32bytes.")
	}
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}
	return &aesPackage{
		Secret: secretKey,
		Block:  block,
	}, nil
}

type Mode int

const (
	PaddingPkcs7 Mode = iota
	PaddingAnsiX923
	PaddingIso10126
	PaddingIso97971
	PaddingZeroPadding
	PaddingNoPadding
)

func (a *aesPackage) ToStringBase64() string {
	return base.Base64Encode(a.CurrentCipher)
}

func (a *aesPackage) ToStringHex() string {
	return fmt.Sprintf("%x", a.CurrentCipher)
}

func (a *aesPackage) ToString() string {
	return string(a.CurrentCipher)
}

func (a *aesPackage) EncryptCFB(message []byte) *aesPackage {
	plainText := message

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Println("EncryptCFB error => ", err)
		return a
	}

	stream := cipher.NewCFBEncrypter(a.Block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	a.CurrentCipher = cipherText
	return a
}

func (a *aesPackage) DecryptCFB(cipherText []byte) *aesPackage {
	if len(cipherText) < aes.BlockSize {
		log.Println("DecryptCFB error => ", "Ciphertext block size is too short!")
		return a
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(a.Block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)

	// return a
	a.CurrentCipher = cipherText

	return a
}

// 加密 CBC 数据填充
// iv 如果跟前端配合，一、固定算法， 二、返回 iv 值给前端
func (a *aesPackage) EncryptCBC(message []byte, mode Mode) *aesPackage {
	//AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := a.Block.BlockSize()

	switch mode {
	case PaddingPkcs7:
		message = PKCSPadding(message, blockSize)
	case PaddingZeroPadding:
		message = ZeroPadding(message, blockSize)
	default:
		log.Println("EncryptCBC error:", "mode类型不匹配")
		return a
	}

	iv := a.Secret[:blockSize] //初始向量的长度必须等于块block的长度16字节

	//log.Println("iv:", string(iv))

	blockMode := cipher.NewCBCEncrypter(a.Block, iv)
	crypted := make([]byte, len(message))
	blockMode.CryptBlocks(crypted, message)

	a.CurrentCipher = crypted
	return a
}

// 解密 CBC
func (a *aesPackage) DecryptCBC(cipherText []byte, iv []byte, mode Mode) *aesPackage {
	//AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := a.Block.BlockSize()
	if len(iv) == 0 {
		iv = a.Secret[:blockSize] // /初始向量的长度必须等于块block的长度16字节
	}
	blockMode := cipher.NewCBCDecrypter(a.Block, iv)
	origData := make([]byte, len(cipherText))
	blockMode.CryptBlocks(origData, cipherText)

	switch mode {
	case PaddingPkcs7:
		origData = PKCSUnPadding(origData)
	case PaddingZeroPadding:
		origData = ZeroUnPadding(origData)
	default:
		log.Println("DecryptCBC error:", "mode类型不匹配")
		return a
	}

	a.CurrentCipher = origData

	return a
}
