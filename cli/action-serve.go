package cli

import (
	"log/slog"

	"github.com/gadfly16/geronimo/server"
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
		slog.Info("Running 'serve' command.")
		if err := server.Serve(sdb); err != nil {
			slog.Error("Serving failed. Exiting.", "error", err.Error())
			runtimeErr = true
		}
	},
}
