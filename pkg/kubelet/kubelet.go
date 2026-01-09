package kubelet

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"superminikube/pkg/api"
	"superminikube/pkg/apiserver/watch"
	"superminikube/pkg/client"
	"superminikube/pkg/kubelet/runtime"
	yaml "superminikube/pkg/spec" // TODO: another bandaid fix until i finish cleaning up
)

// this function will need a context passed to it, once it becomes concurrent
func (k *Kubelet) reconcilePod(event watch.WatchEvent) {
	switch event.EventType {
	case watch.Add:
		slog.Info("creating pod with spec... on node...") // contexts that i plan to add
		cid, err := k.handlePodCreate(*event.Pod.ContainerSpec)
		if err != nil {
			slog.Error("failed to create pod", "err", err)
			return
		}
		// TODO: figure out where i actually want to update the map
		event.Pod.Container.ContainerId = cid
		k.pods[event.Pod.Uid] = event.Pod
	case watch.Delete:
		break
	default:
		slog.Error("Unknown event type")
	}
}

// TODO: move this to PodManager service
// Pod lifecycle sync loop
func (k *Kubelet) syncLoop(ctx context.Context, events <-chan watch.WatchEvent) {
	// block until kubelet receives an event
	// handle event based on type
	for {
		select {
		case <-ctx.Done():
			slog.Info("syncLoop stopped due to context cancellation")
			return
		case event := <-events:
			slog.Debug("Got event", "event", event)
			k.reconcilePod(event)
		}
	}
}

func (k *Kubelet) handlePodDelete(param any) {
	panic("unimplemented")
}

func (k *Kubelet) handlePodCreate(spec api.ContainerSpec) (string, error) {
	slog.Info("Creating pods with spec", "spec", spec)
	// pull image
	// get container opts
	// create container
	// start container
	// set status
	// create pod, then store in map
	err := k.runtime.Pull(spec.Image)
	if err != nil {
		return "", fmt.Errorf("failed to launch pod: %v", err)
	}
	containerOpts, err := yaml.Decode(&spec)
	if err != nil {
		return "", fmt.Errorf("failed to launch pod: %v", err)
	}
	containerId, err := k.runtime.CreateContainer(containerOpts)
	if err != nil {
		return "", fmt.Errorf("failed to launch pod: %v", err)
	}
	err = k.runtime.StartContainer(containerId)
	if err != nil {
		return "", fmt.Errorf("failed to launch pod: %v", err)
	}
	return containerId, nil
}

// Cleanup Kubelet if process killed/stopped
// stops and removes containers running in pods
func (k *Kubelet) Cleanup() []error {
	slog.Info("cleaning up")
	stoppedContainers := make([]string, 0, len(k.pods)) // only store container ids for now
	errs := make([]error, 0)
	for _, p := range k.pods {
		err := k.runtime.StopContainer(p.Container.ContainerId)
		if err != nil {
			err = fmt.Errorf("failed to stop container\nid: %s,\n err: %v", p.Container.ContainerId, err)
			errs = append(errs, err)
			continue
		}
		stoppedContainers = append(stoppedContainers, p.Container.ContainerId) // probably better if you list the containers that FAILED
	}
	slog.Debug("containers stopped", "containers", stoppedContainers)
	return errs
}

func (k *Kubelet) Start(ctx context.Context) error {
	defer k.Cleanup()
	err := k.runtime.Ping()
	if err != nil {
		return fmt.Errorf("Kubelet failed to start: %v", err)
	}

	events, err := k.client.Watch(ctx)
	if err != nil {
		return fmt.Errorf("failed to watch events: %v", err)
	}

	go k.syncLoop(ctx, events)

	<-ctx.Done()
	return nil
}

func NewKubelet(apiServerURL, nodeName string) (*Kubelet, error) {
	rt, err := runtime.NewDockerRuntime()
	if err != nil {
		return nil, fmt.Errorf("failed to create kubelet: %v", err)
	}
	client := client.NewHTTPClient(apiServerURL, nodeName)
	return &Kubelet{
		client:   client,
		runtime:  rt,
		pods:     map[uuid.UUID]api.Pod{},
		nodeName: nodeName,
	}, nil
}

type Kubelet struct {
	client   client.Client // TODO: may come up with better naming convention later
	runtime  runtime.Runtime
	pods     map[uuid.UUID]api.Pod
	nodeName string
}
