package ifile

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

/*
重要使用提示：如何避免图片哈希计算失败
如果你需要对同一个文件既检查类型又计算哈希，请按以下方式操作：
场景 A：传入的是 *os.File
必须在两次读取之间调用 Seek(0, 0) 复位。
f, _ := os.Open("image.jpg")
defer f.Close()

// 1. 检查图片类型 (会读取文件头)
isImg := ifile.IsImage(f, ifile.CheckTypeMimeType)

// 2. 关键：复位文件指针到开头！
f.Seek(0, 0)

// 3. 计算哈希
hash, _ := ifile.FileHashSha256(f)


场景 B：传入的是 io.Reader (如 HTTP 上传)
由于 io.Reader 通常是一次性的，无法 Seek。你需要使用 TeeReader 或者先读取到内存（如果是小图片）：
// 使用 TeeReader 在读取的同时计算哈希，效率最高
var buf bytes.Buffer
tee := io.TeeReader(uploadReader, &buf)

// 此时读取 tee 会同时填充 buf 并在后续可以计算哈希
hash, _ := ifile.FileHashSha256(&buf)


*/

// FileHashSha256 计算 SHA256
func FileHashSha256(reader io.Reader) (string, error) {
	h := sha256.New()
	// io.Copy 默认使用 32KB 缓冲区，能有效防止大文件导致的内存溢出
	if _, err := io.Copy(h, reader); err != nil {
		return "", fmt.Errorf("sha256 copy error: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// FileHashMd5 计算 MD5
func FileHashMd5(reader io.Reader) (string, error) {
	h := md5.New()
	if _, err := io.Copy(h, reader); err != nil {
		return "", fmt.Errorf("md5 copy error: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// FileHashSHA1 计算 SHA1
func FileHashSHA1(reader io.Reader) (string, error) {
	h := sha1.New()
	if _, err := io.Copy(h, reader); err != nil {
		return "", fmt.Errorf("sha1 copy error: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
