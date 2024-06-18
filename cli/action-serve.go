package cli

import (
	"github.com/gadfly16/geronimo/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "runs Geronimo server",
	Long:  `The 'serve' command starts the Geronimo backend server.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Running 'serve' command.")
		if err := server.Serve(s); err != nil {
			log.Error(err.Error())
			runtimeErr = true
		}
	},
}

// Wait for stop signals and quit gently
// signals := make(chan os.Signal, 1)
// signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
// for range signals {
// 	log.Warn("Stopping...")
// 	return nil
// }
