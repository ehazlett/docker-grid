package common

import (
	"github.com/samalba/dockerclient"
)

type (
	NodeData struct {
		NodeId     string                    `json:"node_id,omitempty"`
		Cpus       float64                   `json:"cpus,omitempty"`
		Memory     float64                   `json:"memory,omitempty"`
		Containers []*dockerclient.Container `json:"containers,omitempty"`
		Version    string                    `json:"version,omitempty"`
		IP         string                    `json:"ip,omitempty"`
	}
)
