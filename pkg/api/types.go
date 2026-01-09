package api

import (
	"github.com/google/uuid"
)

type PodSpec struct {
	Container Container
}

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
	// innards
	Spec          PodSpec
}

type Container struct {
	ContainerId string `json:"containerid"`
	Image       string
	Env         map[string]string
	Ports       []Port
	Volumes     []string
}
