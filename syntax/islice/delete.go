package islice

import "slices"

// 介意顺序
func RemoveIntWithOrder(slice []int, s int) []int {
	return append(slice[:s], slice[s+1:]...)
}

// 不介意顺序
func RemoveIntWithoutOrder(s []int, i int) []int {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func RemoveStringWithOrder(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}

func RemoveStringWithoutOrder(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// RemoveElement 不介意顺序 性能高效 替换索引值为最后内容
func RemoveElement[T any](s []T, index int) []T {
	s[index] = s[len(s)-1]
	return s[:len(s)-1]
}

// 字符串的 slice 去空的字符串 使用 ~string 约束，允许所有底层为 string 的类型
func RemoveEmptyString[T ~string](s []T) []T {
	return slices.DeleteFunc(s, func(e T) bool {
		return e == ""
	})
}
