package utils

import (
	"sync"
)

type Map[key comparable, val interface{}] struct {
	mapped map[key]val
	mutex  sync.RWMutex
}

func NewMap[k comparable, v interface{}]() Map[k, v] {
	return Map[k, v]{mapped: map[k]v{}}
}

func NewFromMap[k comparable, v interface{}](m map[k]v) Map[k, v] {
	return Map[k, v]{mapped: m}
}

func (m *Map[key, val]) Get(k key) (val, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	v, ok := m.mapped[k]
	return v, ok
}

func (m *Map[key, val]) Set(k key, v val) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.mapped[k] = v
}

func (m *Map[key, val]) Delete(k key) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.mapped, k)
}
func (m *Map[key, val]) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for eachKey := range m.mapped {
		delete(m.mapped, eachKey)
	}
}

func (m *Map[key, val]) Values() []val {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	arr := make([]val, 0)
	for _, eachVal := range m.mapped {
		arr = append(arr, eachVal)
	}
	return arr
}

func (m *Map[key, val]) Keys() []key {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	arr := make([]key, 0)
	for eachKey := range m.mapped {
		arr = append(arr, eachKey)
	}
	return arr
}
