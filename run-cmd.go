package main

import (
	"flag"

	log "github.com/sirupsen/logrus"
)

func init() {
	commands["run"] = runCommand
}

func runCommand() {
	log.Debug("Running 'run' command.")

	flags := flag.NewFlagSet("run", flag.ExitOnError)

	flags.Parse(flag.Args()[1:])

}
