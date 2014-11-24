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

	Nodes []*NodeData

	NodesById struct {
		Nodes
	}
)

func (n Nodes) Len() int      { return len(n) }
func (n Nodes) Swap(i, j int) { n[i], n[j] = n[j], n[i] }

func (n NodesById) Less(i, j int) bool {
	return n.Nodes[i].NodeId < n.Nodes[j].NodeId
}
