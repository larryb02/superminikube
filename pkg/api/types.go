package api

import (
	"github.com/google/uuid"
)

// type ContainerSpec struct {
// 	Image   string            `yaml:"image"` // required
// 	Env     map[string]string `yaml:"env,omitempty"`
// 	Ports   []Port            `yaml:"ports,omitempty"`   // make this a map or a string of type "hostport:containerport"
// 	Volumes []string          `yaml:"volumes,omitempty"` // only supporting empty volumes at the moment, may also want to move this to Pod level
// }

// TODO
type PodSpec struct {
	Container Container
}

// The specfile
// type Spec struct {
// 	ContainerSpec []ContainerSpec `yaml:"spec"`
// }

type Port struct {
	Hostport      string `yaml:"hostport"`
	Containerport string `yaml:"containerport"`
}

type Pod struct {
	// TODO: create metadata type
	// metadata
	Nodename  string    `json:"nodename"`
	Uid       uuid.UUID `json:"uid"`
	Namespace string    `json:"namespace"` // only default namespace will exist for now
	// TODO: will have to refactor spec and store a PodSpec instead of ContainerSpec
	// This works for now since the only thing a Pod contains is ContainerSpec
	Spec          PodSpec
	// ContainerSpec *ContainerSpec `json:"containerSpec"`
// 	Container     Container
}

type Container struct {
	ContainerId string `json:"containerid"`
	Image       string
	Env         map[string]string
	Ports       []Port
	Volumes     []string
}
