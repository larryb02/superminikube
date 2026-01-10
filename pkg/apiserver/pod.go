package apiserver

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log/slog"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	"superminikube/pkg/api"
)

// TODO: Implement update and append when the need arises

func ListAllNamespacePods(kvstore *redis.Client) ([]api.Pod, error) {
	// scan all keys for pods

	return nil, nil
}

// write tests for crud (get/set/create/delete) of Pod objects via the apiserver
// implement redis store once basic pod lifecycle flow is working
func GetPodByUid(ctx context.Context, nodename string, uid string, kvstore *redis.Client) (api.Pod, error) {
	// TODO: Refactor need to get the pod by /resource/nodename/uid
	slog.Info(fmt.Sprintf("Getting Pod with UID: %s", uid))
	pod, err := kvstore.Get(ctx, fmt.Sprintf("pods/%s/%s", nodename, uid)).Result()
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
func CreatePod(ctx context.Context, nodename string, spec api.PodSpec, kvstore *redis.Client) (api.Pod, error) {
	// takes a decoded spec and creates a Pod object to store in-memory
	var buf bytes.Buffer
	// TODO: make sure uuid is unique
	pod := api.Pod{
		Uid: uuid.New(),
		Nodename: nodename,
		Spec: spec,
	}
	// TODO: shouldn't be encoding in this function
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(pod)
	if err != nil {
		return api.Pod{}, fmt.Errorf("failed to encode pod: %v", err)
	}
	// flatten key into "resource/nodename/identifier"
	err = kvstore.Set(ctx, fmt.Sprintf("pods/%s/%s", nodename, pod.Uid.String()), buf.Bytes(), 0).Err()
	if err != nil {
		return api.Pod{}, fmt.Errorf("failed to store pod: %v", err)
	}
	slog.Info("Created Pod", "pod", pod)
	return pod, nil
}
