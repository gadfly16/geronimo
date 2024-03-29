package cli

import (
	"github.com/gadfly16/geronimo/server"
	log "github.com/sirupsen/logrus"
)

func init() {
	CommandHandlers["serve"] = serveCLI
}

func serveCLI(s server.Settings) error {
	log.Debug("Running 'serve' command.")

	// serveFlags := flag.NewFlagSet("run", flag.ExitOnError)
	// serveFlags.Parse(flag.Args()[1:])

	return server.Serve(s)

	// Wait for stop signals and quit gently
	// signals := make(chan os.Signal, 1)
	// signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	// for range signals {
	// 	log.Warn("Stopping...")
	// 	return nil
	// }
}
