package apiV3

import (
	"strconv"
	"strings"
)

// CompareVersion 比较版本号。v1 > v2 返回 1; v1 == v2 返回 0; v1 < v2 返回 -1
func CompareVersion(v1, v2 string) int {
	// 2026 实践：预处理，去除常见的 'v' 前缀并修剪空格
	s1 := strings.Split(strings.TrimPrefix(strings.TrimSpace(v1), "v"), ".")
	s2 := strings.Split(strings.TrimPrefix(strings.TrimSpace(v2), "v"), ".")

	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	for i := 0; i < maxLen; i++ {
		n1 := parsePart(s1, i)
		n2 := parsePart(s2, i)

		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}

	return 0
}

func parsePart(parts []string, index int) int {
	if index >= len(parts) {
		return 0
	}

	// 2026 健壮性优化：处理可能带后缀的情况，如 "3-rc1"
	raw := parts[index]
	if idx := strings.IndexAny(raw, "-+"); idx != -1 {
		raw = raw[:idx]
	}

	val, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return val
}
