package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/ehazlett/docker-grid/node"
)

var nodeCommand = cli.Command{
	Name:   "node",
	Usage:  "start grid node",
	Action: nodeAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "controller, c",
			Value: "http://127.0.0.1:8080",
			Usage: "URL to controller",
		},
		cli.StringFlag{
			Name:  "docker, d",
			Value: "unix:///var/run/docker.sock",
			Usage: "URL to Docker",
		},
		cli.Float64Flag{
			Name:  "cpus",
			Value: 0.0,
			Usage: "maximum cpus to consume",
		},
		cli.Float64Flag{
			Name:  "memory",
			Value: 0.0,
			Usage: "maximum memory to consume",
		},
		cli.IntFlag{
			Name:  "heartbeat, b",
			Value: 500,
			Usage: "node heartbeat interval (in milliseconds)",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug logging",
		},
	},
}

func waitForInterrupt() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	for _ = range sigChan {
		os.Exit(0)
	}
}

func nodeAction(c *cli.Context) {
	node, err := node.NewNode(c.String("controller"), c.String("docker"), nil, c.Float64("cpus"), c.Float64("memory"), c.Int("heartbeat"), c.Bool("debug"))
	if err != nil {
		log.Fatalf("error connecting to docker: %s", err)
	}

	node.Pulse()
	waitForInterrupt()
}
