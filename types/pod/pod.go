package pod

import (
	"fmt"
	"superminikube/spec"

	"github.com/google/uuid"
)

type PodState int

const (
	PodPending PodState = iota
	PodRunning
	PodFailed
	PodTerminated
)

type Pod struct {
	ContainerSpec *spec.ContainerSpec
	CurrentState  PodState
	UID           uuid.UUID
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
	}, nil
}
