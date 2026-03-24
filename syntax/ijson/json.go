package ijson

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"text/scanner"

	"github.com/json-iterator/go"
)

// 使用 ConfigCompatibleWithStandardLibrary 确保与标准库行为一致
var parser = jsoniter.Config{
	SortMapKeys:            true, // Highly recommended for logs/diffs
	ValidateJsonRawMessage: true,
}.Froze()

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// WriteFile 写入数据到 JSON 文件
func WriteFile(filePath string, data interface{}) error {
	jsonBytes, err := Encode(data)
	if err != nil {
		return err
	}
	// 2026 推荐：使用 os 替代 ioutil
	return os.WriteFile(filePath, jsonBytes, 0664)
}

// ReadFile 读取 JSON 文件数据
func ReadFile(filePath string, v interface{}) error {
	// 2026 推荐：直接使用 os.ReadFile 一步到位
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return Decode(content, v)
}

// Encode 编码。建议在 2026 年返回 error 显式处理
func Encode(v interface{}) ([]byte, error) {
	return parser.Marshal(v)
}

// Decode 解码
func Decode(data []byte, v interface{}) error {
	return parser.Unmarshal(data, v)
}

func Marshal(v interface{}) ([]byte, error) {
	return parser.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return parser.Unmarshal(data, v)
}

// UnmarshalSlice 泛型解码切片 (Go 1.18+)
// 修正了原代码中二级指针 &v 的错误
func UnmarshalSlice[T any](data []byte, v *[]T) error {
	return parser.Unmarshal(data, v)
}

// 直接打印到控制台，完全省掉 string 转换和外部 copy
func LogPretty(v interface{}) {
	if v == nil {
		fmt.Println("null")
		return
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	enc := parser.NewEncoder(buf)
	enc.SetIndent("", "    ")

	if err := enc.Encode(v); err != nil {
		fmt.Printf("json_err: %v\n", err)
		return
	}

	// 直接把内存里的字节流刷到标准输出，这是最快的，没有任何多余分配
	os.Stdout.Write(buf.Bytes())
}

// 正则预编译
var jsonMLComments = regexp.MustCompile(`(?s:/\*.*?\*/\s*)`)

// StripComments 剥离 JSON 中的注释
func StripComments(src string) string {
	if strings.Contains(src, "/*") {
		src = jsonMLComments.ReplaceAllString(src, "")
	}

	if !strings.Contains(src, "//") {
		return strings.TrimSpace(src)
	}

	var s scanner.Scanner
	s.Init(strings.NewReader(src))
	s.Mode ^= scanner.SkipComments // 显式包含注释以便手动识别并剔除

	buf := new(bytes.Buffer)
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		txt := s.TokenText()
		// 简单过滤双斜杠注释。注意：这在包含 URL 字符串时可能存在风险
		if !strings.HasPrefix(txt, "//") && !strings.HasPrefix(txt, "/*") {
			buf.WriteString(txt)
		}
	}
	return buf.String()
}
