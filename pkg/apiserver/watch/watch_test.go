package watch

import (
	"context"
	"fmt"
	"superminikube/pkg/apiserver/store"
	"sync"
	"testing"
	"time"
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

func TestNotify(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func(*WatchService) string
		event    store.StoreEvent
		wantErr  bool
		validate func(*testing.T, chan store.StoreEvent, store.StoreEvent)
	}{
		{
			name: "successful notification",
			setup: func(ws *WatchService) string {
				key := "pod/test-node"
				ws.Watch(key)
				return key
			},
			event: store.StoreEvent{
				Resource: "pod",
				Node:     "test-node",
				Type:     store.EventSet,
			},
			wantErr: false,
			validate: func(t *testing.T, ch chan store.StoreEvent, expected store.StoreEvent) {
				select {
				case received := <-ch:
					if received.Resource != expected.Resource || received.Node != expected.Node || received.Type != expected.Type {
						t.Errorf("received event %+v, expected %+v", received, expected)
					}
				case <-time.After(100 * time.Millisecond):
					t.Error("timeout waiting for event")
				}
			},
		},
		{
			name: "notify non-existent key",
			setup: func(ws *WatchService) string {
				return "pod/nonexistent"
			},
			event: store.StoreEvent{
				Resource: "pod",
				Node:     "nonexistent",
				Type:     store.EventSet,
			},
			wantErr:  true,
			validate: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			ws := New(ctx)
			key := tc.setup(ws)

			var err error
			if tc.validate != nil {
				go func() {
					err = ws.Notify(tc.event)
				}()

				ch, _ := ws.Get(key)
				tc.validate(t, ch, tc.event)

				if (err != nil) != tc.wantErr {
					t.Errorf("Notify() error = %v, wantErr %v", err, tc.wantErr)
				}
			} else {
				err = ws.Notify(tc.event)
				if (err != nil) != tc.wantErr {
					t.Errorf("Notify() error = %v, wantErr %v", err, tc.wantErr)
				}
			}
		})
	}
}

func TestShutdown(t *testing.T) {
	testCases := []struct {
		name         string
		watcherKeys  []string
		expectClosed bool
		expectEmpty  bool
	}{
		{
			name:         "shutdown with multiple watchers",
			watcherKeys:  []string{"pod/node1", "pod/node2", "pod/node3"},
			expectClosed: true,
			expectEmpty:  true,
		},
		{
			name:         "shutdown with single watcher",
			watcherKeys:  []string{"pod/node1"},
			expectClosed: true,
			expectEmpty:  true,
		},
		{
			name:         "shutdown with no watchers",
			watcherKeys:  []string{},
			expectClosed: false,
			expectEmpty:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			ws := New(ctx)

			channels := make([]<-chan store.StoreEvent, 0, len(tc.watcherKeys))
			for _, key := range tc.watcherKeys {
				ch := ws.Watch(key)
				channels = append(channels, ch)
			}

			if len(ws.watchers) != len(tc.watcherKeys) {
				t.Errorf("expected %d watchers before shutdown, got %d", len(tc.watcherKeys), len(ws.watchers))
			}

			ws.Shutdown()

			if tc.expectClosed {
				for i, ch := range channels {
					_, ok := <-ch
					if ok {
						t.Errorf("channel %d should be closed but is still open", i)
					}
				}
			}

			if tc.expectEmpty && len(ws.watchers) != 0 {
				t.Errorf("expected watchers map to be empty, got %d entries", len(ws.watchers))
			}
		})
	}
}
