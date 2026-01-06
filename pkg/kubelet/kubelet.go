package kubelet

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"superminikube/pkg/kubelet/runtime"
	"superminikube/pkg/types/pod"
)

func (k *Kubelet) LaunchPod(p *pod.Pod) error {
	err := k.runtime.Pull(p.ContainerSpec.Image)
	if err != nil {
		return fmt.Errorf("failed to launch pod: %v", err)
	}
	containerOpts, err := p.ContainerSpec.Decode()
	if err != nil {
		return fmt.Errorf("failed to launch pod: %v", err)
	}
	containerId, err := k.runtime.CreateContainer(containerOpts)
	if err != nil {
		return fmt.Errorf("failed to launch pod: %v", err)
	}
	err = k.runtime.StartContainer(containerId)
	if err != nil {
		return fmt.Errorf("failed to launch pod: %v", err)
	}
	p.ContainerId = containerId
	p.CurrentState = pod.PodRunning
	return nil
}

// TODO: Remove
func (k *Kubelet) GetPods() (map[uuid.UUID]*pod.Pod, error) {
	return k.pods, nil // TODO: return a deep copy
}

// Cleanup Kubelet if process killed/stopped
// stops and removes containers running in pods
func (k *Kubelet) Cleanup() []error {
	stoppedContainers := make([]string, 0, len(k.pods)) // only store container ids for now
	errs := make([]error, 0)
	for _, p := range k.pods {
		err := k.runtime.StopContainer(p.ContainerId)
		if err != nil {
			err = fmt.Errorf("failed to stop container\nid: %s,\n err: %v", p.ContainerId, err)
			errs = append(errs, err)
			continue
		}
		stoppedContainers = append(stoppedContainers, p.ContainerId) // probably better if you list the containers that FAILED
	}
	slog.Debug("containers stopped", "containers", stoppedContainers)
	return errs
}

func NewKubelet(rt runtime.Runtime, ctx context.Context) (*Kubelet, error) {
	// sanity check
	err := rt.Ping()
	if err != nil {
		return nil, fmt.Errorf("Kubelet failed to start: %v", err)
	}
	return &Kubelet{
		runtime: rt,
		pods:    map[uuid.UUID]*pod.Pod{},
		ctx:     ctx,
	}, nil
}

type Kubelet struct {
	runtime runtime.Runtime
	pods    map[uuid.UUID]*pod.Pod
	ctx     context.Context
}

// func (k *Kubelet) Apply(specfile string) error {
// 	// for now we're still just loading a specfile, but now we're creating Pods instead of individual containers
// 	// Load spec file
// 	specs, err := spec.CreateSpec(specfile)
// 	if err != nil {
// 		return fmt.Errorf("kubelet: %v", err)
// 	}
// 	var g errgroup.Group
// 	var mu sync.Mutex
// 	for _, spec := range specs.ContainerSpec {
// 		g.Go(func() error {
// 			pod, err := pod.NewPod(&spec)
// 			if err != nil {
// 				return fmt.Errorf("kubelet: %v", err)
// 			}
// 			err = k.LaunchPod(pod)
// 			if err != nil {
// 				return fmt.Errorf("kubelet: %v", err)
// 			}
// 			mu.Lock()
// 			k.pods[pod.UID] = pod
// 			mu.Unlock()
// 			return nil
// 		})
// 	}
// 	return g.Wait()
// }
