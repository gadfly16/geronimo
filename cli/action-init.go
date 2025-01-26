package cli

import (
	"log/slog"
	"time"

	"github.com/gadfly16/geronimo/tree"
	"github.com/spf13/cobra"
)

var (
	logLevelName string
)

func init() {
	initCmd.PersistentFlags().StringVarP(&logLevelName, "log-level", "L", "debug", "logging level")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initializes database and secret keys",
	Long: `The 'init' command initializes all required files in the
			working directory for the application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var ll slog.Level
		var ok bool
		if ll, ok = tree.LogLevelNames[logLevelName]; !ok {
			slog.Error("Unknown log level name.", "levelName", logLevelName)
			return
		}
		rp.LogLevel = int(ll)

		err := tree.InitTree(sdb, rp)
		if err != nil {
			slog.Error("Failed to initialize tree. Exiting.", "error", err.Error())
			return
		}
		slog.Info("Waiting for goroutines to start. TODO")
		time.Sleep(time.Millisecond * 100)
		tree.Tree.Sys.Root.Ask(tree.SystemUser, tree.M_Stop)
		if err := tree.CloseDB(); err != nil {
			slog.Error("State db connection close failed.", "error", err)
		}
		slog.Info("Geronimo initialized.")
	},
}
