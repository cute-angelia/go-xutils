package itxt

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// EnsureObjectUTF8 检查 MinIO 中的文件编码，如果不是 UTF-8 则转换并覆盖
func EnsureObjectUTF8(client *minio.Client, bucket, key string) error {
	ctx := context.Background()

	// 1. 获取原始流
	object, err := client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	// 2. 预读前 2048 字节用于编码检测 (2K 足够检测绝大多数文件)
	header := make([]byte, 2048)
	n, _ := io.ReadAtLeast(object, header, 2048)
	if n == 0 {
		return nil // 空文件无需处理
	}

	// 3. 编码检测逻辑
	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(header[:n])
	if err != nil {
		return fmt.Errorf("failed to detect encoding: %w", err)
	}

	// 如果已经是 UTF-8 且没有 BOM 需求，直接跳过
	if strings.Contains(result.Charset, "UTF-8") {
		return nil
	}

	log.Printf("检测到编码 [%s]，正在转换为 UTF-8: %s/%s", result.Charset, bucket, key)

	// 4. 重置 Object 指针回到开头，准备全量读取
	_, err = object.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to seek object: %w", err)
	}

	// 5. 根据检测结果配置转码器
	var decoder transform.Transformer
	switch result.Charset {
	case "GBK", "GB18030", "GB2312":
		decoder = simplifiedchinese.GBK.NewDecoder()
	case "UTF-16BE":
		decoder = unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()
	case "UTF-16LE":
		decoder = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
	default:
		// 无法确定时，默认尝试 GBK (国内最常见) 并处理可能存在的 BOM
		decoder = simplifiedchinese.GBK.NewDecoder()
	}

	// 使用 BOMOverride 包装：它会自动处理 UTF-8 BOM，如果没 BOM 则使用上面选定的 decoder
	transformer := unicode.BOMOverride(decoder)
	utf8Reader := transform.NewReader(object, transformer)

	// 6. 写入临时文件 (防止内存溢出)
	tempFile, err := os.CreateTemp("", "transcode_*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // 函数结束自动清理
	defer tempFile.Close()

	newSize, err := io.Copy(tempFile, utf8Reader)
	if err != nil {
		return fmt.Errorf("failed to write transcode data: %w", err)
	}

	// 7. 将转换后的内容上传回 MinIO
	_, err = tempFile.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = client.PutObject(ctx, bucket, key, tempFile, newSize, minio.PutObjectOptions{
		ContentType: "text/plain; charset=utf-8",
	})
	if err != nil {
		return fmt.Errorf("failed to upload utf8 file: %w", err)
	}

	log.Printf("成功转换并上传: %d bytes", newSize)
	return nil
}
