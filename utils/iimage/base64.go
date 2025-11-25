package iimage

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/cute-angelia/go-xutils/utils/generator/base"
	"golang.org/x/image/tiff"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// BufferToBase64 accepts a buffer and returns a
// base64 encoded string.
func BufferToBase64(buf bytes.Buffer) string {
	enc := encode(buf.Bytes())
	mime := http.DetectContentType(buf.Bytes())

	return format(enc, mime)
}

// LocalFileToBase64 reads a local file and returns
// the base64 encoded version.
func LocalFileToBase64(fname string) (string, error) {
	var b bytes.Buffer

	fileExists, _ := exists(fname)
	if !fileExists {
		return "", fmt.Errorf("File does not exist\n")
	}

	file, err := os.Open(fname)
	if err != nil {
		return "", fmt.Errorf("Error opening file\n")
	}

	_, err = b.ReadFrom(file)
	if err != nil {
		return "", fmt.Errorf("Error reading file to buffer\n")
	}

	return BufferToBase64(b), nil
}

// Base64ToLocalFile base64 字符串转本地文件
func Base64ToLocalFile(base64Str string, localfile string) error {
	lengthBase64 := len(base64Str)
	if lengthBase64 == 0 {
		return errors.New("base64为空")
	}
	// 1. 从base64中解析出mime类型
	index := strings.IndexByte(base64Str, ',')

	if index == -1 {
		return errors.New("base64为空")
	}

	mime := base64Str[0:index]
	mime = strings.Replace(mime, "data:", "", 1)
	mime = strings.Replace(mime, ";base64", "", 1)

	base64Data := base64Str[index+1:]

	// 移除干扰
	base64Data = base.ParseB64String(base64Data, false)

	// 4. 解码base64字符串获取图片数据
	imgData, err := base.Base64Decode(base64Data)
	if err != nil {
		log.Println(err)
		return err
	}

	// 5. 构造文件名及文件路径
	dir := filepath.Dir(localfile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// 6. 将图片数据写入文件
	out, err := os.OpenFile(localfile, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer out.Close()

	// 7. 编码图片信息
	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return errors.New("base64不合法" + err.Error())
	}

	switch mime {
	case "image/jpeg", "image/jpg", "image/pjpeg":
		jpeg.Encode(out, img, &jpeg.Options{Quality: 100})
	case "image/png":
		png.Encode(out, img)
	case "image/gif":
		gif.Encode(out, img, &gif.Options{})
	case "image/tiff":
		tiff.Encode(out, img, &tiff.Options{})
	default:
		return errors.New("base64不合法 mime")
	}
	return nil
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// encode is our main function for
// base64 encoding a passed []byte
func encode(bin []byte) []byte {
	e64 := base64.StdEncoding
	maxEncLen := e64.EncodedLen(len(bin))
	encBuf := make([]byte, maxEncLen)
	e64.Encode(encBuf, bin)
	return encBuf
}

// format is an abstraction of the mime switch to create the
// acceptable base64 string needed for browsers.
func format(enc []byte, mime string) string {
	switch mime {
	case "image/gif", "image/jpeg", "image/pjpeg", "image/png", "image/tiff":
		return fmt.Sprintf("data:%s;base64,%s", mime, enc)
	default:
	}
	return fmt.Sprintf("data:image/png;base64,%s", enc)
}
