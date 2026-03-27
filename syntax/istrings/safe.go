package istrings

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// SanitizeName 替换文件名或文件夹名中的非法字符
// replacement 为替换字符，通常用 "_" 或 "-"
func SanitizeName(name string, replacement string) string {
	if replacement == "" {
		replacement = "_"
	}

	// Windows 非法字符: \ / : * ? " < > |
	illegalChars := `\/:*?"<>|`

	var builder strings.Builder
	for _, r := range name {
		switch {
		// 1. 处理非法符号、null 字节、控制字符
		case strings.ContainsRune(illegalChars, r) || r == 0 || unicode.IsControl(r):
			builder.WriteString(replacement)

		// 2. 核心新增：过滤 Emoji 及其他特殊符号
		// unicode.So 涵盖了绝大多数 Emoji (如 🍎, ⭐, 😃)
		// unicode.IsSymbol 涵盖了数学符号、货币符号等 (如 ©, ®, ±)
		case unicode.IsSymbol(r) || IsEmoji(r):
			builder.WriteString(replacement)

		default:
			builder.WriteRune(r)
		}
	}

	result := builder.String()

	// 去除首尾空格和点（Windows 规范）
	result = strings.TrimSpace(result)
	// 注意：如果是处理目录名，Trim 点号是正确的；如果是带后缀的文件名，这里需要小心
	result = strings.Trim(result, ".")

	// 检查处理后是否为空
	if result == "" {
		// 生成 4 位随机数后缀，防止批量更名冲突
		// 这里使用微秒取模，简单高效
		randSuffix := time.Now().UnixMicro() % 10000
		result = fmt.Sprintf("unnamed_%04d", randSuffix)
	}

	// Windows 保留名称检查
	reservedNames := map[string]bool{
		"CON": true, "PRN": true, "AUX": true, "NUL": true,
		"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
		"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
	}

	upper := strings.ToUpper(result)
	baseName := upper
	if dotIdx := strings.LastIndex(upper, "."); dotIdx != -1 {
		baseName = upper[:dotIdx]
	}

	if reservedNames[baseName] {
		result = replacement + result
	}

	return result
}
