package main

import (
	"net"
	"strings"

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
		cli.StringFlag{
			Name:  "ip, i",
			Value: "",
			Usage: "machine IP",
		},
		cli.BoolFlag{
			Name:  "grid-containers, g",
			Usage: "show only grid containers",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug logging",
		},
	},
}

func nodeAction(c *cli.Context) {
	if c.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	}
	nodeIp := c.String("ip")
	// if no host ip is specified, attempt to detect
	if nodeIp == "" {
		// get listening IP and check if running in "bridged" mode
		// if so, exit and tell to run in "host" mode
		// this is so we can properly report the IP of the host in port listings
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			log.Fatalf("unable to get network interfaces: %s", err)
		}

		log.Debugf("detecting machine ip...")
		for _, addr := range addrs {
			i, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}
			ip := i.String()
			if i.To4() == nil {
				log.Debugf("skipping ipv6 address: %s", ip)
				continue
			}
			switch {
			case ip == "127.0.0.1":
				continue
			case strings.Index(ip, "172.17") != -1:
				continue
			case strings.Index(ip, "fe80") != -1:
				continue
			case ip == "::1":
				continue
			default:
				nodeIp = ip
				break
			}
		}
		if nodeIp == "" {
			log.Fatalf("unable to run node: unable to detect machine IP -- use --net=host if you are running in a container")
		}

		log.Debugf("detected machine ip: %s", nodeIp)

	}

	node, err := node.NewNode(c.String("controller"), c.String("docker"), nil, c.Float64("cpus"), c.Float64("memory"), c.Int("heartbeat"), nodeIp, c.Bool("grid-containers"), c.Bool("debug"))
	if err != nil {
		log.Fatalf("error connecting to docker: %s", err)
	}

	node.Run()

	waitForInterrupt()
}
