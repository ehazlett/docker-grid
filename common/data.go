package common

import (
	"github.com/samalba/dockerclient"
)

type (
	NodeData struct {
		NodeId     string
		Cpus       float64
		Memory     float64
		Containers []dockerclient.Container
	}
)
