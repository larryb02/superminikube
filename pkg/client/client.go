package client

import (
	"context"

	"github.com/google/uuid"

	"superminikube/pkg/apiserver/store"
)

// Client provides an interface for interacting with the API server
type Client interface {
	// Resource operations
	Get(ctx context.Context, resource string, id uuid.UUID) ([]byte, error)
	List(ctx context.Context, resource string) ([]byte, error)
	Update(ctx context.Context, resource string, id uuid.UUID, data []byte) error

	// Watch for events from the control plane
	Watch(ctx context.Context) (<-chan store.StoreEvent, error)

	// Health check
	Ping(ctx context.Context) error
}
