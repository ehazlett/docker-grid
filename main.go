package main

import (
	"os"

	"github.com/codegangsta/cli"
)

var ()

func main() {
	app := cli.NewApp()
	app.Name = "grid"
	app.Usage = "docker cluster grid"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{}
	app.Commands = []cli.Command{
		nodeCommand,
		controllerCommand,
	}

	app.Run(os.Args)
}
