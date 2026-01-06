package apiserver

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log/slog"

	"superminikube/pkg/apiserver/store"
	"superminikube/pkg/spec"
	podv2 "superminikube/pkg/types/pod/v2"
)

// TODO: Implement update and append when the need arises

func ListAllNamespacePods(kvstore store.Store) ([]podv2.Pod, error) {
	// scan all keys for pods

	return nil, nil
}

// write tests for crud (get/set/create/delete) of Pod objects via the apiserver
// implement redis store once basic pod lifecycle flow is working
func GetPodByUid(nodename string, uid string, kvstore store.Store) (podv2.Pod, error) {
	// TODO: Refactor need to get the pod by /resource/nodename/uid
	slog.Info(fmt.Sprintf("Getting Pod with UID: %s", uid))
	pod, err := kvstore.Get(fmt.Sprintf("pods/%s/%s", nodename, uid))
	if err != nil {
		return podv2.Pod{}, fmt.Errorf("failed to get pod from store: %v", err)
	}
	var p podv2.Pod
	decoder := gob.NewDecoder(bytes.NewReader(pod))
	err = decoder.Decode(&p)
	if err != nil {
		return podv2.Pod{}, fmt.Errorf("failed to decode pod: %v", err)
	}
	return p, nil
}

// NOTE: Return type could be of type CreatePodResponse in the future
func CreatePod(nodename string, spec *spec.ContainerSpec, kvstore store.Store) (podv2.Pod, error) {
	// takes a decoded spec and creates a Pod object to store in-memory
	var buf bytes.Buffer
	// TODO: make sure uuid is unique
	pod := podv2.New(nodename, spec)
	// TODO: shouldn't be encoding in this function
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(pod)
	if err != nil {
		return podv2.Pod{}, fmt.Errorf("failed to encode pod: %v", err)
	}
	// TODO: make sure get, set, and delete support multiple types
	// also consider that we should probably create an entry for the node's environment on some sort of register event
	// flatten key into "resource/nodename/identifier" -> at least for pods
	err = kvstore.Set(fmt.Sprintf("pods/%s/%s", nodename, pod.Uid.String()), buf.Bytes())
	if err != nil {
		return podv2.Pod{}, fmt.Errorf("failed to store pod: %v", err)
	}
	slog.Info("Created Pod", "pod", pod)
	return *pod, nil
}
