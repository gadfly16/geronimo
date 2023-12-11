package cli

import (
	"flag"

	"github.com/gadfly16/geronimo/server"
	log "github.com/sirupsen/logrus"
)

func init() {
	Commands["init"] = initCommand
}

func initCommand(settings server.Settings) error {
	log.Debug("Running 'init' command.")

	initFlags := flag.NewFlagSet("init", flag.ExitOnError)
	initFlags.Parse(flag.Args()[1:])

	return server.CreateDB(settings)
}
