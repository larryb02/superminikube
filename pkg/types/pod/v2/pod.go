package v2

import (
	"superminikube/pkg/spec"

	"github.com/google/uuid"
)

func New(nodename string, containerSpec *spec.ContainerSpec) *Pod {
	uuid := uuid.New()
	return &Pod{
		Nodename:      nodename,
		Uid:           uuid,
		ContainerSpec: containerSpec,
	}
}

// NOTE: Defined here instead of /types because making the change there
// will break the node agent code that imports /types/pod
// need to refactor the node agent before moving this to types - maybe this will only belong in apiserver :shrug:
type Pod struct {
	Nodename string    `json:"nodename"`
	Uid      uuid.UUID `json:"uid"`
	// TODO: will have to refactor spec and store a PodSpec instead of ContainerSpec
	// This works for now since the only thing a Pod contains is ContainerSpec
	ContainerSpec *spec.ContainerSpec `json:"containerSpec"`
}
