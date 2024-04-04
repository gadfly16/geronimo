package main

import (
	"github.com/gadfly16/geronimo/server"
	log "github.com/sirupsen/logrus"
)

func init() {
	parser.AddCommand(
		"serve",
		"run Geronimo server",
		"starts the Geronimo backend server",
		&serveOptions{})
}

type serveOptions struct{}

func (serveOpts *serveOptions) Execute(args []string) error {
	s.Init()
	log.Debug("Running 'serve' command.")

	return server.Serve(s)

	// Wait for stop signals and quit gently
	// signals := make(chan os.Signal, 1)
	// signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	// for range signals {
	// 	log.Warn("Stopping...")
	// 	return nil
	// }
}
