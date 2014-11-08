package controller

import (
	"encoding/gob"
	"encoding/json"
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/docker-grid/common"
	"github.com/ehazlett/docker-grid/utils/datastore"
	"github.com/samalba/dockerclient"
)

type (
	Node struct {
		Id     string
		Cpus   float64
		Memory float64
	}

	Controller struct {
		Addr      string
		ApiAddr   string
		Nodes     []*Node
		TTL       int
		datastore *datastore.Datastore
	}
)

func NewController(addr string, apiAddr string, ttl int, enableDebug bool) (*Controller, error) {
	ds, err := datastore.New(time.Millisecond * time.Duration(ttl))
	if err != nil {
		return nil, err
	}
	controller := &Controller{
		Addr:      addr,
		ApiAddr:   apiAddr,
		TTL:       ttl,
		datastore: ds,
	}
	if enableDebug {
		log.SetLevel(log.DebugLevel)
	}
	controller.init()
	return controller, nil
}

func (c *Controller) init() {
	gob.Register(common.NodeData{})
	gob.Register([]dockerclient.Container{})
}

func (c *Controller) handleConnection(conn net.Conn) {
	log.Debugf("heartbeat from %s", conn.RemoteAddr())
	dec := gob.NewDecoder(conn)
	data := &common.NodeData{}
	if err := dec.Decode(data); err != nil {
		log.Warnf("error decoding data: %s", err)
		return
	}

	// update datastore
	c.datastore.Set(data.NodeId, data)
}

func (c *Controller) Run() error {
	listener, err := net.Listen("tcp", c.Addr)
	if err != nil {
		return err
	}

	http.HandleFunc("/", c.ApiIndex)
	http.HandleFunc("/containers/json", c.ApiListContainers)
	http.HandleFunc("/v1.15/containers/json", c.ApiListContainers)
	go http.ListenAndServe(c.ApiAddr, nil)

	log.Infof("grid controller listening on %s", c.Addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Warnf("client connection error: %s", err)
			continue
		}
		go c.handleConnection(conn)
	}
}

// API emulation
func (c *Controller) ApiIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("docker grid controller\n"))
}

func (c *Controller) ApiListContainers(w http.ResponseWriter, r *http.Request) {
	var containers []dockerclient.Container
	for _, v := range c.datastore.Items() {
		containers = append(containers, v.Data.(*common.NodeData).Containers...)
	}
	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(containers); err != nil {
		log.Warnf("error encoding container response: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
