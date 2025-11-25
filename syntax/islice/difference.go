package islice

// Difference 返回在 a 中但不在 b 中的元素（差集）
func Difference[T comparable](a, b []T) []T {
	bMap := make(map[T]struct{}, len(b))
	for _, item := range b {
		bMap[item] = struct{}{}
	}

	result := make([]T, 0)
	for _, item := range a {
		if _, exists := bMap[item]; !exists {
			result = append(result, item)
		}
	}
	return result
}

// Union 返回 a 和 b 的并集（去重）
func Union[T comparable](a, b []T) []T {
	unionMap := make(map[T]struct{})

	for _, item := range a {
		unionMap[item] = struct{}{}
	}
	for _, item := range b {
		unionMap[item] = struct{}{}
	}

	result := make([]T, 0, len(unionMap))
	for item := range unionMap {
		result = append(result, item)
	}
	return result
}

// Intersection 返回 a 和 b 的交集
func Intersection[T comparable](a, b []T) []T {
	bMap := make(map[T]struct{}, len(b))
	for _, item := range b {
		bMap[item] = struct{}{}
	}

	result := make([]T, 0)
	seen := make(map[T]struct{})
	for _, item := range a {
		if _, exists := bMap[item]; exists {
			if _, duplicate := seen[item]; !duplicate {
				result = append(result, item)
				seen[item] = struct{}{}
			}
		}
	}
	return result
}

// SymmetricDifference 返回对称差集（在 a 或 b 中，但不在两者共同部分）
func SymmetricDifference[T comparable](a, b []T) []T {
	aMap := make(map[T]struct{}, len(a))
	bMap := make(map[T]struct{}, len(b))

	for _, item := range a {
		aMap[item] = struct{}{}
	}
	for _, item := range b {
		bMap[item] = struct{}{}
	}

	result := make([]T, 0)
	for item := range aMap {
		if _, exists := bMap[item]; !exists {
			result = append(result, item)
		}
	}
	for item := range bMap {
		if _, exists := aMap[item]; !exists {
			result = append(result, item)
		}
	}
	return result
}
