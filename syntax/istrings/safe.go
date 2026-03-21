package istrings

import (
	"strings"
	"unicode"
)

// SanitizeName 替换文件名或文件夹名中的非法字符
// replacement 为替换字符，通常用 "_" 或 "-"
func SanitizeName(name string, replacement string) string {
	if replacement == "" {
		replacement = "_"
	}

	// Windows 非法字符: \ / : * ? " < > |
	// Linux/macOS 非法字符: / 和 null 字节
	// 统一处理，兼容三个平台
	illegalChars := `\/:*?"<>|`

	var builder strings.Builder
	for _, r := range name {
		switch {
		case strings.ContainsRune(illegalChars, r): // 非法符号
			builder.WriteString(replacement)
		case r == 0: // null 字节
			builder.WriteString(replacement)
		case unicode.IsControl(r): // 控制字符 (ASCII 0-31, 127)
			builder.WriteString(replacement)
		default:
			builder.WriteRune(r)
		}
	}

	result := builder.String()

	// 去除首尾空格和点（Windows 不允许）
	result = strings.TrimSpace(result)
	result = strings.Trim(result, ".")

	// Windows 保留名称（不区分大小写）
	reservedNames := map[string]bool{
		"CON": true, "PRN": true, "AUX": true, "NUL": true,
		"COM1": true, "COM2": true, "COM3": true, "COM4": true,
		"COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
		"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true,
		"LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
	}

	// 检查去掉扩展名后的名称是否为保留名
	upper := strings.ToUpper(result)
	dotIdx := strings.LastIndex(upper, ".")
	baseName := upper
	if dotIdx != -1 {
		baseName = upper[:dotIdx]
	}
	if reservedNames[baseName] {
		result = replacement + result
	}

	// 如果处理后为空，给一个默认名
	if result == "" {
		result = "unnamed"
	}

	return result
}
