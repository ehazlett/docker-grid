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
	"strings"
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

func NewNode(controllerUrl string, dockerUrl string, tlsConfig *tls.Config, cpus float64, memory float64, heartbeatInterval int, enableDebug bool) (*Node, error) {
	if enableDebug {
		log.SetLevel(log.DebugLevel)
	}

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
	allContainers, err := node.ListContainers(false)
	if err != nil {
		log.Warnf("error listing containers: %s", err)
	}

	var containers []*dockerclient.Container
	// filter non-grid containers
	for _, cnt := range allContainers {
		c := cnt
		info, err := node.client.InspectContainer(c.Id)
		if err != nil {
			log.Warnf("unable to inspect container: %s", c.Id)
			continue
		}

		for _, e := range info.Config.Env {
			k := strings.Split(e, "=")
			if k[0] == "DOCKER_GRID" {
				containers = append(containers, &c)
				break
			}
		}
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

func (node *Node) checkQueue() {
	resp, err := node.doRequest("/grid/queue/next", "GET", 200, nil)
	if err != nil {
		log.Warnf("error checking queue: %s", err)
		return
	}

	var job common.Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		log.Warnf("error decoding job: %s", err)
		return
	}

	if job.Id != "" {
		log.Infof("processing job: id=%s image=%s", job.Id, job.ContainerConfig.Image)
		// TODO: inject "grid" env var
		cntCfg := job.ContainerConfig
		if cntCfg.Env == nil {
			env := []string{"DOCKER_GRID=true"}
			cntCfg.Env = env
		} else {
			cntCfg.Env = append(cntCfg.Env, "DOCKER_GRID=true")
		}
		containerId, err := node.createContainer(cntCfg)
		if err != nil {
			log.Warnf("error creating container: %s")
		}
		result := &common.JobResult{
			JobId:       job.Id,
			NodeId:      node.Id,
			ContainerId: containerId,
		}
		hostCfg := job.ContainerConfig.HostConfig
		if err := node.client.StartContainer(containerId, &hostCfg); err != nil {
			log.Warnf("error starting container: %s", err)
		}

		info, err := node.client.InspectContainer(containerId)
		if err != nil {
			log.Warnf("error inspecting container: %s", err)
		}

		result.ContainerInfo = info

		if err != nil {
			result.Warnings = []string{err.Error()}
		}

		b, err := json.Marshal(result)
		if err != nil {
			log.Fatalf("error marshaling job result: %s", err)
		}
		if _, err := node.doRequest("/grid/queue/result", "POST", 200, b); err != nil {
			log.Warnf("error sending job result: %s", err)
		}
	}
}

func (node *Node) createContainer(config *dockerclient.ContainerConfig) (string, error) {
	id, err := node.client.CreateContainer(config, "")
	if err != nil {
		if err.Error() == "Not found" {
			log.Debugf("attempting to pull image %s: ", config.Image)
			if imgErr := node.client.PullImage(config.Image); imgErr != nil {
				log.Warnf("error pulling image: %s", imgErr)
				return "", imgErr
			}
			id, err = node.client.CreateContainer(config, "")
		} else {
			log.Warnf("error creating container: %s", err)
		}
	}
	return id, err
}

func (node *Node) Pulse() {
	ticker := time.NewTicker(time.Millisecond * time.Duration(node.heartbeatInterval))

	go func() {
		for _ = range ticker.C {
			node.sendContainers()
			node.checkQueue()
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
