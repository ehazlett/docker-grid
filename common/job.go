package common

import (
	"time"

	"github.com/samalba/dockerclient"
)

type (
	Job struct {
		Id              string                        `json:"id,omitempty"`
		Date            time.Time                     `json:"date,omitempty"`
		ContainerName   string                        `json:"container_name"`
		ContainerConfig *dockerclient.ContainerConfig `json:"container_config,omitempty"`
	}

	JobResult struct {
		JobId         string                      `json:"id,omitempty"`
		NodeId        string                      `json:"node_id,omitempty"`
		ContainerId   string                      `json:"container_id"`
		ContainerInfo *dockerclient.ContainerInfo `json:"container_info,omitempty"`
		Warnings      []string                    `json:"warnings"`
	}
)
