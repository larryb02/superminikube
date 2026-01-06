package watch

import (
	"context"
	"fmt"
	"log/slog"
	"superminikube/pkg/apiserver/store"
	"sync"
)

func (ws *WatchService) Get(key string) (chan store.StoreEvent, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	val, ok := ws.watchers[key]
	if !ok {
		return nil, fmt.Errorf("key %v doesn't exist", key)
	}
	return val, nil
}

func (ws *WatchService) Set(key string, value chan store.StoreEvent) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	ws.watchers[key] = value
	return nil
}

func (ws *WatchService) Shutdown() {
	// slog.Debug("waiting for goroutines")
	// ws.wg.Wait()
	slog.Debug("cleaning up channels")
	for k := range ws.watchers {
		// TODO: Create a cleanup method that does this
		// this way can cleanup on client disconnects or do it in bulk on shutdown
		close(ws.watchers[k])
		delete(ws.watchers, k)
	}
}

// TODO: Move notify outside of this package, belongs in store
// will take chan as argument
// Notify watch service of a mutation event
func (ws *WatchService) Notify(ev store.StoreEvent) error {
	key := fmt.Sprintf("%s/%s", ev.Resource, ev.Node)
	slog.Debug("notifying watcher", "watcher", ws.watchers[key], "event", ev, "key", key)
	ws.mu.Lock()
	defer ws.mu.Unlock()
	_, ok := ws.watchers[key]
	if !ok {
		return fmt.Errorf("failed to notify key: %s not found", key)
	}
	ws.watchers[key] <- ev
	return nil
}

// // TODO: Delete
// func (ws *WatchService) WatchLoop(key string) <-chan store.StoreEvent {
// 	out := make(chan store.StoreEvent)
// 	// TODO: Eventually key should be fully parameterized
// 	// once other resources get introduced
// 	key = fmt.Sprintf("pod/%s", key)
// 	ws.watchers[key] = out
// 	ws.wg.Go(func() {
// 		slog.Info("starting watch loop")
// 		// ch := ws.Watch(key)
// 		// for {
// 		// select {
// 		// keys should always be resource/nodename
// 		// we are watching for changes in a nodes environment
// 		// watch may become more generalized than this
// 		// case ev := <-ch:
// 		// 	slog.Info("received kvstore event", "event", ev)
// 		// 	out <- ev
// 		<-ws.ctx.Done()
// 		slog.Info("closing watch connection")
// 		// return
// 		// }
// 		// }
// 	})
// 	return out
// }

func (ws *WatchService) Watch(key string) <-chan store.StoreEvent {
	// make a channel for the key add it to map of channels
	// TODO: Avoid making duplicate channels
	ev := make(chan store.StoreEvent)
	err := ws.Set(key, ev)
	if err != nil {
		// TODO: return error
		slog.Error("failed to add channel to watchers")
		return nil
	}
	return ev
}

func New(parentCtx context.Context) *WatchService {
	ctx, cancel := context.WithCancel(parentCtx)
	return &WatchService{
		watchers: make(Watchers),
		ctx:      ctx,
		cancel:   cancel,
	}
}

type Watchers map[string]chan store.StoreEvent
type WatchService struct {
	watchers Watchers
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.RWMutex
}
