package watch

import (
	"fmt"
	"log/slog"
	"sync"

	"superminikube/pkg/api"
)

type Service interface {
	Watch(string) <-chan WatchEvent
	Notify(WatchEvent) error
}

// should probably make getter and setter private
func (ws *WatchService) Get(key string) (chan WatchEvent, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	val, ok := ws.watchers[key]
	if !ok {
		return nil, fmt.Errorf("key %v doesn't exist", key)
	}
	return val, nil
}

func (ws *WatchService) Set(key string, value chan WatchEvent) error {
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

// TODO: Move notify outside of this package, most likely belongs elsewhere
// will take chan as argument
// Notify watch service of a mutation event
func (ws *WatchService) Notify(ev WatchEvent) error {
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

func (ws *WatchService) Watch(key string) <-chan WatchEvent {
	// make a channel for the key add it to map of channels
	// TODO: Avoid making duplicate channels
	ev := make(chan WatchEvent)
	err := ws.Set(key, ev)
	if err != nil {
		// TODO: return error
		slog.Error("failed to add channel to watchers")
		return nil
	}
	return ev
}

func NewService() *WatchService {
	return &WatchService{
		watchers: make(Watchers),
	}
}

type WatchEvent struct {
	EventType Event
	// these fields are tentative
	Resource string
	Node     string
	// TODO: will need to pass a general 'object' here
	// will be an interface that every object implements
	// or will have a field that designates it's type
	Pod api.Pod
}

// TODO: Add events type here
const (
	Add Event = iota
	Delete
)

type Event int
type Watchers map[string]chan WatchEvent
type WatchService struct {
	watchers Watchers
	mu       sync.RWMutex
}
