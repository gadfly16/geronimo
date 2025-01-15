package cli

import (
	"log/slog"
	"time"

	"github.com/gadfly16/geronimo/core"
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
		if ll, ok = core.LogLevelNames[logLevelName]; !ok {
			slog.Error("Unknown log level name.", "levelName", logLevelName)
			return
		}
		rp.LogLevel = int(ll)

		err := core.InitTree(sdb, rp)
		if err != nil {
			slog.Error("Failed to initialize tree. Exiting.", "error", err.Error())
			return
		}
		slog.Info("Waiting for goroutines to start. TODO")
		time.Sleep(time.Millisecond * 100)
		core.Tree.Sys.Root.Ask(core.StopMsgKind, core.SystemUser)
		if err := core.CloseDB(); err != nil {
			slog.Error("State db connection close failed.", "error", err)
		}
		slog.Info("Geronimo initialized.")
	},
}
