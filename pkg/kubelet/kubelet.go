package kubelet

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/moby/moby/api/types/jsonstream"
	mobyclient "github.com/moby/moby/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"superminikube/pkg/api"
	"superminikube/pkg/apiserver/watch"
	"superminikube/pkg/client"
	"superminikube/pkg/kubelet/runtime"
)

func (k *Kubelet) handlePodEvent(ctx context.Context, event watch.WatchEvent) {
	switch event.EventType {
	case watch.Add:
		slog.Info("creating pod with spec... on node...")
		cid, err := k.handlePodCreate(ctx, event.Pod.Spec)
		if err != nil {
			slog.Error("failed to create pod", "err", err)
			return
		}
		event.Pod.Spec.Container.ContainerId = cid
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
	for {
		select {
		case <-ctx.Done():
			slog.Info("syncLoop stopped due to context cancellation")
			return
		case event := <-events:
			slog.Debug("Got event", "event", event)
			k.handlePodEvent(ctx, event)
		}
	}
}

func (k *Kubelet) handlePodDelete(param any) {
	panic("unimplemented")
}

// Creates container then returns container id
func (k *Kubelet) handlePodCreate(ctx context.Context, spec api.PodSpec) (string, error) {
	slog.Info("Creating pods with spec", "spec", spec)
	// pull image
	pullOpts := mobyclient.ImagePullOptions{
		Platforms: []ocispec.Platform{{Architecture: "amd64", OS: "linux"}},
	}
	slog.Info("Attempting to pull", "image", spec.Container.Image)
	resp, err := k.containerruntime.ImagePull(ctx, spec.Container.Image, pullOpts)
	if err != nil {
		return "", fmt.Errorf("failed to pull image: %v", err)
	}
	var pullErrs []*jsonstream.Error
	for m := range resp.JSONMessages(ctx) {
		if m.Error != nil {
			pullErrs = append(pullErrs, m.Error)
		} else {
			slog.Info("Status:", "status", m.Status)
		}
	}
	if len(pullErrs) > 0 {
		return "", fmt.Errorf("failed to pull image: %v", pullErrs)
	}
	containerOpts, err := runtime.PodSpecToCreateContainerOpts(spec)
	if err != nil {
		return "", fmt.Errorf("failed to create container opts: %v", err)
	}
	createRes, err := k.containerruntime.ContainerCreate(ctx, containerOpts)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %v", err)
	}
	slog.Info("Created", "container", createRes.ID)
	// start container
	_, err = k.containerruntime.ContainerStart(ctx, createRes.ID, mobyclient.ContainerStartOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to start container: %v", err)
	}
	slog.Info("Started", "container", createRes.ID)
	return createRes.ID, nil
}

func (k *Kubelet) Shutdown(ctx context.Context) {
	removedContainers := make([]string, 0, len(k.pods))
	errs := make([]error, 0)
	for _, p := range k.pods {
		err := k.CleanupPod(ctx, p)
		if err != nil {
			err = fmt.Errorf("id: %s\terr: %v", p.Spec.Container.ContainerId, err)
			errs = append(errs, err)
			continue
		}
		removedContainers = append(removedContainers, p.Spec.Container.ContainerId)
	}
	if len(errs) > 0 {
		slog.Error("failed to remove containers", "containers", errs)
	}
	slog.Debug("containers removed", "containers", removedContainers)
}

// Cleanup Kubelet if process killed/stopped
// stops and removes containers running in pods
func (k *Kubelet) CleanupPod(ctx context.Context, p api.Pod) error {
	ctx = context.WithoutCancel(ctx)
	cid := p.Spec.Container.ContainerId
	slog.Info("removing container", "containerid", cid)
	_, err := k.containerruntime.ContainerRemove(ctx, cid, mobyclient.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})
	if err != nil {
		return fmt.Errorf("failed to remove container in pod: %v\nerr: %v", p.Uid, err)
	}
	return nil
}

func (k *Kubelet) Start(ctx context.Context) error {
	defer k.Shutdown(ctx)
	_, err := k.containerruntime.Ping(ctx, mobyclient.PingOptions{})
	if err != nil {
		return fmt.Errorf("Kubelet failed to start: %v", err)
	}
	slog.Info("Successfully pinged Docker")
	events, err := k.client.Watch(ctx)
	if err != nil {
		return fmt.Errorf("failed to watch events: %v", err)
	}
	go k.syncLoop(ctx, events)
	<-ctx.Done()
	return nil
}

func NewKubelet(apiServerURL, nodeName string) (*Kubelet, error) {
	rt, err := mobyclient.New(mobyclient.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubelet: %v", err)
	}
	client := client.NewHTTPClient(apiServerURL, nodeName)
	return &Kubelet{
		client:   client,
		containerruntime:  rt,
		pods:     map[uuid.UUID]api.Pod{},
		nodeName: nodeName,
	}, nil
}

type Kubelet struct {
	client   client.Client
	containerruntime  *mobyclient.Client
	pods     map[uuid.UUID]api.Pod
	nodeName string
}
