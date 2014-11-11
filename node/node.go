package node

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
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
	return node, nil
}

func (node *Node) buildUrl(path string) string {
	return fmt.Sprintf("%s%s", node.controllerUrl, path)
}

func (node *Node) doRequest(path string, method string, expectedStatus int, b []byte) (*http.Response, error) {
	url := node.buildUrl(path)
	buf := bytes.NewBuffer(b)
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err

	}
	req.Header.Set("User-Agent", "grid-node")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err

	}

	if resp.StatusCode != expectedStatus {
		c, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err

		}
		return resp, errors.New(string(c))

	}
	return resp, nil

}

func (node *Node) sendContainers() {
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

	b, err := json.Marshal(d)
	if err != nil {
		log.Fatalf("error marshaling containers: %s", err)
	}

	if _, err := node.doRequest(fmt.Sprintf("/grid/nodes/%s/update", node.Id), "POST", 200, b); err != nil {
		log.Warnf("error sending heartbeat: %s", err)
	}
}

func (node *Node) Pulse() {
	ticker := time.NewTicker(time.Millisecond * time.Duration(node.heartbeatInterval))

	go func() {
		for _ = range ticker.C {
			node.sendContainers()
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
