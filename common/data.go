package common

import (
	"time"

	"github.com/samalba/dockerclient"
)

type (
	NodeData struct {
		NodeId     string                   `json:"node_id,omitempty"`
		Cpus       float64                  `json:"cpus,omitempty"`
		Memory     float64                  `json:"memory,omitempty"`
		Containers []dockerclient.Container `json:"containers,omitempty"`
	}

	Job struct {
		Id              string                        `json:"id,omitempty"`
		Date            time.Time                     `json:"date,omitempty"`
		ContainerConfig *dockerclient.ContainerConfig `json:"container_config,omitempty"`
	}
)
