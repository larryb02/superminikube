package pod

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log/slog"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	"superminikube/pkg/api"
	"superminikube/pkg/apiserver/watch"
)

type Service interface {
	GetPodByUid(ctx context.Context, nodename, uid string) (api.Pod, error)
	ListAllNamespacePods(ctx context.Context) ([]api.Pod, error)
	CreatePod(ctx context.Context, nodename string, spec api.PodSpec) (api.Pod, error)
}

type PodService struct {
	kvstore      *redis.Client
	watchService *watch.WatchService
}

func NewService(kvstore *redis.Client, watchService *watch.WatchService) *PodService {
	return &PodService{
		kvstore:      kvstore,
		watchService: watchService,
	}
}

// TODO: Implement update and append when the need arises

func (s *PodService) ListAllNamespacePods(ctx context.Context) ([]api.Pod, error) {
	// TODO
	return nil, nil
}

// write tests for crud (get/set/create/delete) of Pod objects via the apiserver
// implement redis store once basic pod lifecycle flow is working
func (s *PodService) GetPodByUid(ctx context.Context, nodename, uid string) (api.Pod, error) {
	// TODO: Refactor need to get the pod by /resource/nodename/uid
	slog.Info(fmt.Sprintf("Getting Pod with UID: %s", uid))
	pod, err := s.kvstore.Get(ctx, fmt.Sprintf("pods/%s/%s", nodename, uid)).Result()
	if err != nil {
		return api.Pod{}, fmt.Errorf("failed to get pod from store: %v", err)
	}
	var p api.Pod
	// TODO: there is probably a better way to decode
	decoder := gob.NewDecoder(bytes.NewReader([]byte(pod)))
	err = decoder.Decode(&p)
	if err != nil {
		return api.Pod{}, fmt.Errorf("failed to decode pod: %v", err)
	}
	return p, nil
}

// NOTE: Return type could be of type CreatePodResponse in the future
// TODO: nodename will not be a parameter here, scheduler will decide where pod goes
func (s *PodService) CreatePod(ctx context.Context, nodename string, spec api.PodSpec) (api.Pod, error) {
	var buf bytes.Buffer
	// TODO: make sure uuid is unique
	pod := api.Pod{
		Uid:      uuid.New(),
		Nodename: nodename,
		Spec:     spec,
	}
	// TODO: shouldn't be encoding in this function
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(pod)
	if err != nil {
		return api.Pod{}, fmt.Errorf("failed to encode pod: %v", err)
	}
	// flatten key into "resource/nodename/identifier"
	err = s.kvstore.Set(ctx, fmt.Sprintf("pods/%s/%s", nodename, pod.Uid.String()), buf.Bytes(), 0).Err()
	if err != nil {
		return api.Pod{}, fmt.Errorf("failed to store pod: %v", err)
	}
	slog.Info("Created Pod", "pod", pod)

	// Notify watch service
	err = s.watchService.Notify(watch.WatchEvent{
		EventType: watch.Add,
		Resource:  "pod",
		Node:      nodename,
		Pod:       pod,
	})
	if err != nil {
		slog.Warn("failed to notify watcher", "error", err)
	}

	return pod, nil
}
