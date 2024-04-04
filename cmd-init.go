package main

import (
	"github.com/gadfly16/geronimo/server"
	log "github.com/sirupsen/logrus"
)

func init() {
	parser.AddCommand(
		"init",
		"init database and secret keys",
		"command initializes all required files to serve Geronimo",
		&initOptions{})
}

type initOptions struct{}

func (initOpts *initOptions) Execute(args []string) error {
	s.Init()
	log.Debug("Running 'init' command.")
	return server.Init(s)
}
