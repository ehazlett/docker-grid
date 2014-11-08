package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/ehazlett/docker-grid/controller"
)

var controllerCommand = cli.Command{
	Name:   "controller",
	Usage:  "start grid controller",
	Action: controllerAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "listen, l",
			Value: ":8080",
			Usage: "controller listen address",
		},
		cli.StringFlag{
			Name:  "api-listen, a",
			Value: ":8081",
			Usage: "docker api listen address",
		},
		cli.IntFlag{
			Name:  "ttl, t",
			Value: 500,
			Usage: "node ttl (in ms)",
		},
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "enable debug logging",
		},
	},
}

func controllerAction(c *cli.Context) {
	controller, err := controller.NewController(c.String("listen"), c.String("api-listen"), c.Int("ttl"), c.Bool("debug"))
	if err != nil {
		log.Fatalf("error creating controller: %s", err)
	}
	if err := controller.Run(); err != nil {
		log.Fatalf("error starting controller: %s", err)
	}
}
