package imap

import "sync"

// SafeMap 采用泛型设计，优化了双缓冲逻辑
type SafeMap[K comparable, V any] struct {
	lock  sync.RWMutex
	count int // 记录删除次数或操作阈值
	data  map[K]V
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		data: make(map[K]V),
	}
}

func (m *SafeMap[K, V]) Set(key K, value V) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.data[key] = value
}

func (m *SafeMap[K, V]) Get(key K) (V, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	val, ok := m.data[key]
	return val, ok
}

func (m *SafeMap[K, V]) Del(key K) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.data[key]; ok {
		delete(m.data, key)
		m.count++
	}

	// 2026 推荐实践：当删除次数过多时，通过重建 Map 来真正释放内存
	// 阈值可根据业务调整，例如 10000
	if m.count > 10000 {
		m.rebuild()
	}
}

// rebuild 核心：通过创建新 map 并拷贝来解决 Issue 20135
func (m *SafeMap[K, V]) rebuild() {
	newMap := make(map[K]V, len(m.data))
	for k, v := range m.data {
		newMap[k] = v
	}
	m.data = newMap
	m.count = 0
}

func (m *SafeMap[K, V]) Range(f func(key K, val V) bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	for k, v := range m.data {
		if !f(k, v) {
			break
		}
	}
}
