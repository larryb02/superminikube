package apiserver

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log/slog"

	"superminikube/pkg/api"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// TODO: Implement update and append when the need arises

func ListAllNamespacePods(kvstore *redis.Client) ([]api.Pod, error) {
	// scan all keys for pods

	return nil, nil
}

// write tests for crud (get/set/create/delete) of Pod objects via the apiserver
// implement redis store once basic pod lifecycle flow is working
func GetPodByUid(nodename string, uid string, kvstore *redis.Client) (api.Pod, error) {
	// TODO: Refactor need to get the pod by /resource/nodename/uid
	slog.Info(fmt.Sprintf("Getting Pod with UID: %s", uid))
	pod, err := kvstore.Get(context.TODO(), fmt.Sprintf("pods/%s/%s", nodename, uid)).Result()
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
func CreatePod(nodename string, spec *api.ContainerSpec, kvstore *redis.Client) (api.Pod, error) {
	// takes a decoded spec and creates a Pod object to store in-memory
	var buf bytes.Buffer
	// TODO: make sure uuid is unique
	pod := api.Pod{
		Uid: uuid.New(),
		Nodename: nodename,
		ContainerSpec: spec,
	}
	// TODO: shouldn't be encoding in this function
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(pod)
	if err != nil {
		return api.Pod{}, fmt.Errorf("failed to encode pod: %v", err)
	}
	// flatten key into "resource/nodename/identifier"
	err = kvstore.Set(context.TODO(), fmt.Sprintf("pods/%s/%s", nodename, pod.Uid.String()), buf.Bytes(), 0).Err() // TODO: pass ctx into each of this functions...
	if err != nil {
		return api.Pod{}, fmt.Errorf("failed to store pod: %v", err)
	}
	slog.Info("Created Pod", "pod", pod)
	return pod, nil
}
