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
	"github.com/gorilla/mux"
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

	r := mux.NewRouter()
	r.HandleFunc("/", c.apiIndex)
	r.HandleFunc("/grid/nodes", c.apiNodeList)
	r.HandleFunc("/grid/nodes/{nodeId}", c.apiNodeDetails)
	r.HandleFunc("/containers/json", c.apiListContainers)
	r.HandleFunc("/v1.15/containers/json", c.apiListContainers)
	http.Handle("/", r)
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

// API
func (c *Controller) apiIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("docker grid controller\n"))
}

func (c *Controller) apiNodeList(w http.ResponseWriter, r *http.Request) {
	data := c.datastore.Items()
	var nodes []*common.NodeData
	for _, v := range data {
		n := &common.NodeData{
			NodeId: v.Data.(*common.NodeData).NodeId,
			Cpus:   v.Data.(*common.NodeData).Cpus,
			Memory: v.Data.(*common.NodeData).Memory,
		}
		nodes = append(nodes, n)
	}
	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(nodes); err != nil {
		log.Warnf("error encoding node list: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Controller) apiNodeDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeId := vars["nodeId"]
	d, err := c.datastore.Get(nodeId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(d); err != nil {
		log.Warnf("error encoding node details: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Docker API compatibility
func (c *Controller) apiListContainers(w http.ResponseWriter, r *http.Request) {
	var containers []dockerclient.Container
	for _, v := range c.datastore.Items() {
		containers = append(containers, v.Data.(*common.NodeData).Containers...)
	}
	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(containers); err != nil {
		log.Warnf("error encoding container response: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
