package cli

import (
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	"github.com/gadfly16/geronimo/msg"
	"github.com/gadfly16/geronimo/node"
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
		if ll, ok = node.LogLevelNames[logLevelName]; !ok {
			slog.Error("Unknown log level name.", "levelName", logLevelName)
			return
		}
		rp.LogLevel = int(ll)

		if err := node.InitDb(sdb); err != nil {
			slog.Error("Failed to create db. Exiting.", "error", err.Error())
			return
		}
		if err := node.ConnectDB(sdb); err != nil {
			slog.Error("Failed to connect to db. Exiting.", "error", err.Error())
			return
		}

		if err := node.InitRootNode(&rp); err != nil {
			slog.Error("Failed to create root node. Exiting.", "error", err.Error())
			return
		}
		r := node.Tree.Root.Ask(
			&msg.Msg{
				Kind: msg.CreateKind,
				Payload: &node.GroupNode{
					Head: &node.Head{
						Name: "Users",
						Kind: node.GroupKind,
					},
				},
			})
		if r.Kind == msg.ErrorKind {
			slog.Error("User group creation failed. Exiting!", "error", r.ErrorMsg())
		}
		slog.Info("Waiting for goroutines to start. TODO")
		time.Sleep(time.Millisecond * 100)
		node.Tree.Root.Ask(msg.Stop)
		if err := node.CloseDB(); err != nil {
			slog.Error("State db connection close failed.", "error", err)
		}
		slog.Info("Geronimo initialized.")
	},
}
