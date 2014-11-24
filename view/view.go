package view

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/docker-grid/common"
	"github.com/olekukonko/tablewriter"
	"github.com/wsxiaoys/terminal/color"
)

type View struct {
	controllerUrl   string
	refreshInterval int
}

func NewView(controllerUrl string, refreshInterval int) (*View, error) {
	return &View{
		controllerUrl:   controllerUrl,
		refreshInterval: refreshInterval,
	}, nil
}

func (v *View) buildUrl(path string) string {
	return fmt.Sprintf("%s%s", v.controllerUrl, path)
}

func (v *View) doRequest(path string, method string, expectedStatus int, b []byte) (*http.Response, error) {
	url := v.buildUrl(path)
	buf := bytes.NewBuffer(b)
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err

	}
	req.Header.Set("User-Agent", "grid-view")

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

func (v *View) Start() {

	ticker := time.NewTicker(time.Second * time.Duration(v.refreshInterval))

	go func() {
		for _ = range ticker.C {
			v.refresh()
		}
	}()

}

func (v *View) getNodes() ([]*common.NodeData, error) {
	resp, err := v.doRequest("/grid/nodes", "GET", 200, nil)
	if err != nil {
		log.Warnf("error getting node list: %s", err)
		return nil, err
	}

	nodeData := []*common.NodeData{}
	if err := json.NewDecoder(resp.Body).Decode(&nodeData); err != nil {
		log.Warnf("error decoding node data: %s", err)
		return nil, err
	}

	return nodeData, nil
}

func (v *View) refresh() {
	nodes, err := v.getNodes()
	if err != nil {
		log.Fatalf("unable to get nodes: %s", err)
	}

	sort.Sort(common.NodesById{nodes})

	fmt.Print("\033[2J")
	fmt.Print("\033[H")
	fmt.Print("|")
	color.Printf("@b Docker Grid: %s\n", v.controllerUrl)
	fmt.Print("|\n")

	if len(nodes) == 0 {
		fmt.Println("| ----- No connected nodes -----")
		fmt.Println("|")
	} else {
		t := tablewriter.NewWriter(os.Stdout)
		t.SetHeader([]string{"", "ID", "CPUs", "Memory", "Version", "IP", "CONTAINERS"})

		for i, node := range nodes {
			cpus := fmt.Sprintf("%.2f", node.Cpus)
			memory := fmt.Sprintf("%.2f", node.Memory)
			if node.Cpus == 0.0 {
				cpus = ""
			}
			if node.Memory == 0.0 {
				memory = ""
			}
			t.Append([]string{fmt.Sprintf("%d", i), node.NodeId, cpus, memory, node.Version, node.IP, fmt.Sprintf("%d", len(node.Containers))})
		}

		t.Render()
	}
}
