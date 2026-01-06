package watch

import (
	"context"
	"fmt"
	"superminikube/pkg/apiserver/store"
	"sync"
	"testing"
)

// hmm probably won't test this
// func TestRegisterWatcher(t *testing.T) {
// 	ctx := context.Background()
// 	ws := New(ctx)
// 	_ = ws
// 	t.Errorf("failed to register watcher")
// }

func TestGet(t *testing.T) {
	ch := make(chan store.StoreEvent)
	ws := WatchService{
		watchers: Watchers{
			"test": ch,
		},
	}
	testCases := []struct {
		key      string
		expected chan store.StoreEvent
		wantErr  bool
	}{
		{"test", ch, false},        // existing key
		{"nonexistent", nil, true}, // missing key
	}

	for _, tc := range testCases {
		val, err := ws.Get(tc.key)
		if (err != nil) != tc.wantErr {
			t.Errorf("Get(%q) error = %v, want %v", tc.key, err, tc.wantErr)
			continue
		}

		if val != tc.expected {
			t.Errorf("Get(%q) = %v, expected %v", tc.key, val, tc.expected)
		}
	}
}

func TestSet(t *testing.T) {
	ctx := context.Background()
	ws := New(ctx)
	testCases := []struct {
		key     string
		value   chan store.StoreEvent
		wantErr bool
	}{
		{"test", nil, false},
		{"", make(chan store.StoreEvent), true},
	}
	for _, tc := range testCases {
		err := ws.Set(tc.key, tc.value)
		if (err != nil) != tc.wantErr {
			t.Errorf("Set(%q) err = %v", tc.key, err)
		}
	}
}

func TestSetGetConcurrently(t *testing.T) {
	ctx := context.Background()
	ws := New(ctx)

	const numGoroutines = 50
	const numOpsPerGoroutine = 100

	var wg sync.WaitGroup

	for g := 0; g < numGoroutines; g++ {
		wg.Go(func() {
			defer wg.Done()
			for i := 0; i < numOpsPerGoroutine; i++ {
				key := fmt.Sprintf("key-%d-%d", g, i)
				ch := make(chan store.StoreEvent)
				if err := ws.Set(key, ch); err != nil {
					t.Errorf("Set(%q) failed: %v", key, err)
				}

				got, err := ws.Get(key)
				if err != nil {
					t.Errorf("Get(%q) failed: %v", key, err)
				}
				if got != ch {
					t.Errorf("Get(%q) = %v, expected %v", key, got, ch)
				}
			}
		})
	}
	wg.Wait()
}
