// TODO: Delete this file
package store

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("not found")

type MemStore struct {
	mu     sync.RWMutex
	items  map[string][]byte
	closed bool
}

// NewMemStore creates an empty in-memory store.
func New() *MemStore {
	return &MemStore{
		items: make(map[string][]byte),
	}
}
func (m *MemStore) ensureKind(kind string) {
	if _, ok := m.items[kind]; !ok {
		m.items[kind] = make([]byte, 0)
	}
}

func (m *MemStore) Append(key string, value any) (error) {
	return nil
}

func (m *MemStore) Watch(key string) <-chan StoreEvent {
	return nil
}

func (m *MemStore) Set(key string, value any) error {
	if key == "" {
		return errors.New("key required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return errors.New("store closed")
	}
	m.ensureKind(key)
	// store a copy
	b := make([]byte, len(value.([]byte)))
	copy(b, value.([]byte))
	m.items[key] = b
	return nil
}

func (m *MemStore) Get(key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.closed {
		return nil, errors.New("store closed")
	}
	data, ok := m.items[key]
	if !ok {
		return nil, ErrNotFound
	}
	// return a copy
	b := make([]byte, len(data))
	copy(b, data)
	return b, nil
}

func (m *MemStore) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return errors.New("store closed")
	}
	_, ok := m.items[key]
	if !ok {
		return ErrNotFound
	}
	delete(m.items, key)
	return nil
}

// // Convenience helpers for Pods to keep existing call patterns easy to use.
// func (m *MemStore) PutPod(p *v2.Pod) error {
// 	if p == nil {
// 		return errors.New("pod is nil")
// 	}
// 	data, err := json.Marshal(p)
// 	if err != nil {
// 		return err
// 	}
// 	return m.Put("pod", p.Uid.String(), data)
// }

// func (m *MemStore) GetPod(uid string) (*v2.Pod, error) {
// 	data, err := m.Get("pod", uid)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var p v2.Pod
// 	if err := json.Unmarshal(data, &p); err != nil {
// 		return nil, err
// 	}
// 	return &p, nil
// }

// func (m *MemStore) DeletePod(uid string) error {
// 	return m.Delete("pod", uid)
// }

// func (m *MemStore) ListPods() ([]*v2.Pod, error) {
// 	dataList, err := m.List("pod")
// 	if err != nil {
// 		return nil, err
// 	}
// 	res := make([]*v2.Pod, 0, len(dataList))
// 	for _, d := range dataList {
// 		var p v2.Pod
// 		if err := json.Unmarshal(d, &p); err != nil {
// 			return nil, err
// 		}
// 		res = append(res, &p)
// 	}
// 	return res, nil
// }

func (m *MemStore) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil
	}
	m.closed = true
	m.items = nil
	return nil
}
