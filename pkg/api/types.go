package api

import (
	"github.com/google/uuid"
)

type ContainerSpec struct {
	Image   string            `yaml:"image"` // required
	Env     map[string]string `yaml:"env,omitempty"`
	Ports   []Port            `yaml:"ports,omitempty"`   // make this a map or a string of type "hostport:containerport"
	Volumes []string          `yaml:"volumes,omitempty"` // only supporting empty volumes at the moment, may also want to move this to Pod level
}

// TODO
type PodSpec struct {
	Name string
}

// The specfile
type Spec struct {
	ContainerSpec []ContainerSpec `yaml:"spec"`
}

type Port struct {
	Hostport      string `yaml:"hostport"`
	Containerport string `yaml:"containerport"`
}

type Pod struct {
	// these two items are metadata and every resource will have metadata hint hint...
	Nodename string    `json:"nodename"`
	Uid      uuid.UUID `json:"uid"`
	// TODO: will have to refactor spec and store a PodSpec instead of ContainerSpec
	// This works for now since the only thing a Pod contains is ContainerSpec
	ContainerSpec *ContainerSpec `json:"containerSpec"`
	Container     Container
}

type Container struct {
	ContainerId string `json:"containerid"`
}
