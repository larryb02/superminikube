package pod

import (
	"fmt"

	"github.com/google/uuid"
	"superminikube/spec"
)

type PodState int

const (
	// TODO: Add a nillish value
	PodPending PodState = iota
	PodRunning
	PodFailed
	PodTerminated
)

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
		ContainerId: "",
	}, nil
}
