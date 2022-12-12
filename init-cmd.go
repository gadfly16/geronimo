package main

import (
	"flag"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func init() {
	commands["init"] = initCommand
}

func initCommand() {
	log.Debug("Running 'init' command.")

	initFlags := flag.NewFlagSet("init", flag.ExitOnError)
	initFlags.Parse(flag.Args()[1:])

	createDB()
}
