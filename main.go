package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/codegangsta/cli"
)

func waitForInterrupt() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	for _ = range sigChan {
		os.Exit(0)
	}
}

const VERSION = "0.0.4"

func main() {
	app := cli.NewApp()
	app.Name = "grid"
	app.Usage = "docker cluster grid"
	app.Version = VERSION
	app.Flags = []cli.Flag{}
	app.Commands = []cli.Command{
		nodeCommand,
		controllerCommand,
		viewCommand,
	}

	app.Run(os.Args)
}
