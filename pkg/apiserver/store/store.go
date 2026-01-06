package store

// Package store provides an abstraction for persisting cluster resources
// (Pods, etc.). Implementations can plug in etcd, redis or an in-memory
// store. A small in-memory implementation is provided for testing and local
// development.

import ()

type StoreEvent struct {
	Type     EventType // Created, Updated, Deleted
	Resource string    // "pod"
	Node     string
	// Key      string // TODO: This or the actual data is needed to perform syncing of state
	// Value    []byte    // optional
}

type EventType int

const (
	EventSet EventType = iota
	EventDelete
)

type Store interface {
	Append(key string, value any) error
	Set(key string, value any) error
	Get(key string) ([]byte, error)
	Delete(key string) error
}
