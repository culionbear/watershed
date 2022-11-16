package mutex

import (
	"errors"
	"reflect"
	"sync"
)

type Value[T comparable] struct {
	m sync.RWMutex
	v T
}

func New[T comparable](value T) (*Value[T], error) {
	t := reflect.TypeOf(value)
	if t.Kind() == reflect.Pointer {
		return nil, errors.New("value is a pointer")
	}
	return &Value[T]{
		v: value,
	}, nil
}

func (m *Value[T]) Load() T {
	m.m.RLock()
	defer m.m.RUnlock()
	return m.v
}

func (m *Value[T]) Store(value T) {
	m.m.Lock()
	defer m.m.Unlock()
	m.v = value
}

func (m *Value[T]) Compare(value T) bool {
	m.m.RLock()
	defer m.m.RUnlock()
	return m.v == value
}

func (m *Value[T]) Do(f func(...T)) {
	m.m.Lock()
	defer m.m.Unlock()
	f(m.v)
}
