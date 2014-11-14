package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/ehazlett/docker-grid/view"
)

var viewCommand = cli.Command{
	Name:   "view",
	Usage:  "start grid viewer",
	Action: viewAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "controller, c",
			Value: "http://127.0.0.1:8080",
			Usage: "URL to controller",
		},
		cli.IntFlag{
			Name:  "refresh, r",
			Value: 1,
			Usage: "view refresh interval (in seconds)",
		},
	},
}

func viewAction(c *cli.Context) {

	view, err := view.NewView(c.String("controller"), c.Int("refresh"))
	if err != nil {
		log.Fatalf("error connecting to docker: %s", err)
	}

	view.Start()

	waitForInterrupt()
}
