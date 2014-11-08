package node

import (
	"crypto/tls"
	"encoding/gob"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/docker-grid/common"
	"github.com/samalba/dockerclient"
	"github.com/twinj/uuid"
)

type (
	Node struct {
		Id                string
		client            *dockerclient.DockerClient
		conn              *net.Conn
		controllerUrl     string
		heartbeatInterval int
		Cpus              float64
		Memory            float64
	}
)

func NewNode(controllerUrl string, dockerUrl string, tlsConfig *tls.Config, cpus float64, memory float64, heartbeatInterval int) (*Node, error) {
	u := uuid.NewV4()
	id := uuid.Formatter(u, uuid.CleanHyphen)

	client, err := dockerclient.NewDockerClient(dockerUrl, tlsConfig)
	if err != nil {
		return nil, err
	}

	node := &Node{
		Id:                id,
		client:            client,
		controllerUrl:     controllerUrl,
		heartbeatInterval: heartbeatInterval,
		Cpus:              cpus,
		Memory:            memory,
	}
	node.init()
	return node, nil
}

func (node *Node) init() {
	gob.Register(common.NodeData{})
	gob.Register([]dockerclient.Container{})
}

func (node *Node) Send(data interface{}) error {
	c, err := net.Dial("tcp", node.controllerUrl)
	if err != nil {
		return err
	}

	enc := gob.NewEncoder(c)
	if err := enc.Encode(data); err != nil {
		return err
	}

	return nil
}

func (node *Node) Pulse() {
	ticker := time.NewTicker(time.Millisecond * time.Duration(node.heartbeatInterval))

	go func() {
		for _ = range ticker.C {
			containers, err := node.ListContainers(false)
			if err != nil {
				log.Warnf("error listing containers: %s", err)
			}

			d := &common.NodeData{
				NodeId:     node.Id,
				Cpus:       node.Cpus,
				Memory:     node.Memory,
				Containers: containers,
			}

			if err := node.Send(d); err != nil {
				log.Warnf("error sending data: %s", err)
			}
		}
	}()

	log.Infof("node started: id=%s cpus=%.2f memory=%.2f heartbeat=%dms", node.Id, node.Cpus, node.Memory, node.heartbeatInterval)
}

func (node *Node) ListContainers(all bool) ([]dockerclient.Container, error) {
	containers, err := node.client.ListContainers(all)
	if err != nil {
		return []dockerclient.Container{}, err
	}
	// TODO: filter grid containers
	return containers, nil
}
