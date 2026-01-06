package pod

import (
	"fmt"

	"superminikube/pkg/spec"

	"github.com/google/uuid"
)

type PodState int

const (
	// TODO: Add a nillish value
	PodPending PodState = iota
	PodRunning
	PodFailed
	PodTerminated
)

// TODO: Create container struct
type Pod struct {
	ContainerSpec *spec.ContainerSpec
	CurrentState  PodState
	UID           uuid.UUID
	ContainerId   string // NOTE: a full struct for containers may be useful
}

func NewPod(spec *spec.ContainerSpec) (*Pod, error) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}
	return &Pod{
		ContainerSpec: spec,
		CurrentState:  PodPending,
		UID:           uuid,
		ContainerId:   "",
	}, nil
}
