package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"code.google.com/p/go-uuid/uuid"
	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/docker-grid/common"
	"github.com/ehazlett/docker-grid/utils/datastore"
	fifo "github.com/foize/go.fifo"
	"github.com/gorilla/mux"
	"github.com/samalba/dockerclient"
)

const VERSION = "0.0.2"

type (
	Node struct {
		Id     string
		Cpus   float64
		Memory float64
	}

	Controller struct {
		Addr               string
		Nodes              []*Node
		TTL                int
		datastore          *datastore.Datastore
		jobResultDatastore *datastore.Datastore
		queue              *fifo.Queue
	}
)

func NewController(addr string, ttl int, enableDebug bool) (*Controller, error) {
	ds, err := datastore.New(time.Millisecond * time.Duration(ttl))
	if err != nil {
		return nil, err
	}
	jobResultDs, err := datastore.New(time.Millisecond * time.Duration(ttl))
	if err != nil {
		return nil, err
	}
	queue := fifo.NewQueue()
	controller := &Controller{
		Addr:               addr,
		TTL:                ttl,
		datastore:          ds,
		jobResultDatastore: jobResultDs,
		queue:              queue,
	}
	if enableDebug {
		log.SetLevel(log.DebugLevel)
	}
	return controller, nil
}

func (c *Controller) ListContainers() []*dockerclient.Container {
	var containers []*dockerclient.Container
	for _, v := range c.datastore.Items() {
		containers = append(containers, v.Data.(*common.NodeData).Containers...)
	}
	return containers
}

func (c *Controller) Run() error {
	r := mux.NewRouter()
	r.HandleFunc("/", c.apiIndex).Methods("GET")
	r.HandleFunc("/grid/nodes", c.apiNodeList).Methods("GET")
	r.HandleFunc("/grid/nodes/{nodeId}", c.apiNodeDetails).Methods("GET")
	r.HandleFunc("/grid/queue/next", c.apiQueueNext).Methods("GET")
	r.HandleFunc("/grid/queue/result", c.apiQueueResult).Methods("POST")
	r.HandleFunc("/grid/nodes/{nodeId}/update", c.apiNodeUpdate).Methods("POST")
	r.HandleFunc("/{apiVersion}/containers/json", c.apiListContainers).Methods("GET")
	r.HandleFunc("/{apiVersion}/containers/create", c.apiCreateContainer).Methods("POST")
	r.HandleFunc("/{apiVersion}/containers/{containerId}/attach", c.apiAttachContainer).Methods("POST")
	r.HandleFunc("/{apiVersion}/containers/{containerId}/start", c.apiStartContainer).Methods("POST")
	r.HandleFunc("/{apiVersion}/containers/{containerId}/wait", c.apiWaitContainer).Methods("POST")
	r.HandleFunc("/{apiVersion}/containers/{containerId}/json", c.apiContainerJson).Methods("GET")
	r.HandleFunc("/{apiVersion}/containers/{containerId}", c.apiDeleteContainer).Methods("DELETE")
	http.Handle("/", r)

	log.Infof("grid controller started: version=%s port=%s", VERSION, c.Addr)

	return http.ListenAndServe(c.Addr, c.logRequest(http.DefaultServeMux))
}

func (c *Controller) logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

// API
func (c *Controller) apiIndex(w http.ResponseWriter, r *http.Request) {
	v := fmt.Sprintf("docker grid controller %s\n", VERSION)
	w.Write([]byte(v))
}

func (c *Controller) apiNodeUpdate(w http.ResponseWriter, r *http.Request) {
	data := &common.NodeData{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// update datastore
	c.datastore.Set(data.NodeId, data)
	w.WriteHeader(http.StatusOK)
}

func (c *Controller) apiNodeList(w http.ResponseWriter, r *http.Request) {
	data := c.datastore.Items()
	var nodes []*common.NodeData
	for _, v := range data {
		n := &common.NodeData{
			NodeId:  v.Data.(*common.NodeData).NodeId,
			Cpus:    v.Data.(*common.NodeData).Cpus,
			Memory:  v.Data.(*common.NodeData).Memory,
			Version: v.Data.(*common.NodeData).Version,
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
		return
	}
	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(d); err != nil {
		log.Warnf("error encoding node details: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Controller) apiQueueNext(w http.ResponseWriter, r *http.Request) {
	job := c.queue.Next()
	if job == nil {
		job = &common.Job{}
	}

	if job.(*common.Job).Id != "" {
		log.Infof("sending job: id=%s image=%s node=%s", job.(*common.Job).Id, job.(*common.Job).ContainerConfig.Image, r.RemoteAddr)
	}

	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(job); err != nil {
		log.Warnf("error encoding job: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Controller) apiQueueResult(w http.ResponseWriter, r *http.Request) {
	result := &common.JobResult{}
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.jobResultDatastore.Set(result.JobId, result)
	log.Infof("received job result: %s", result.JobId)
	w.WriteHeader(http.StatusOK)
}

// Docker API compatibility
func (c *Controller) apiListContainers(w http.ResponseWriter, r *http.Request) {
	containers := c.ListContainers()
	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(containers); err != nil {
		log.Warnf("error encoding container response: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Controller) apiCreateContainer(w http.ResponseWriter, r *http.Request) {
	var containerConfig dockerclient.ContainerConfig

	if err := json.NewDecoder(r.Body).Decode(&containerConfig); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// queue job
	job := &common.Job{
		Id:              uuid.New(),
		Date:            time.Now(),
		ContainerConfig: &containerConfig,
	}
	c.queue.Add(job)

	log.Infof("queue job: id=%s image=%s", job.Id, job.ContainerConfig.Image)

	log.Debugf("pending jobs: %d", c.queue.Len())

	// wait for response

	resp := &dockerclient.RespContainersCreate{
		Id:       "pending",
		Warnings: []string{},
	}

	for {
		rs, err := c.jobResultDatastore.Get(job.Id)
		if err != nil {
			time.Sleep(500)
			continue
		}
		result := rs.(*datastore.Item).Data.(*common.JobResult)
		resp.Id = result.ContainerId
		resp.Warnings = result.Warnings
		break
	}

	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Warnf("error encoding container response: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Controller) apiAttachContainer(w http.ResponseWriter, r *http.Request) {
	// HACK: hijack response
}

func (c *Controller) apiStartContainer(w http.ResponseWriter, r *http.Request) {
	// HACK: hijack response
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) apiWaitContainer(w http.ResponseWriter, r *http.Request) {
	// HACK: hijack response
	resp := &common.WaitResponse{
		StatusCode: 0, // TODO: get actual status code from JobResult
	}
	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Warnf("error encoding container wait response: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Controller) apiContainerJson(w http.ResponseWriter, r *http.Request) {
	// HACK: hijack response
	vars := mux.Vars(r)
	containerId := vars["containerId"]
	containerInfo := &dockerclient.ContainerInfo{}
	for _, v := range c.jobResultDatastore.Items() {
		result := v.Data.(*common.JobResult)
		if strings.Index(result.ContainerId, containerId) == 0 {
			containerInfo = result.ContainerInfo
		}
	}
	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(containerInfo); err != nil {
		log.Warnf("error encoding container config: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Controller) apiDeleteContainer(w http.ResponseWriter, r *http.Request) {
	// HACK: hijack response
	w.WriteHeader(http.StatusForbidden)
}
