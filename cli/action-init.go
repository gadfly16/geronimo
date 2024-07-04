package cli

import (
	"errors"
	"log/slog"
	"time"

	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

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

		if err := InitDb(sdb); err != nil {
			slog.Error("Failed to create db. Exiting.", "error", err.Error())
			return
		}
		if err := node.ConnectDB(sdb); err != nil {
			slog.Error("Failed to connect to db. Exiting.", "error", err.Error())
			return
		}

		root := &node.RootNode{
			Head: &node.Head{
				Name: "Root",
				Kind: node.RootKind,
			},
			Parms: &rp,
		}
		if err := root.Init(); err != nil {
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
			slog.Error("User group creation failed. Exiting!", "error", r.Error())
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

func InitDb(path string) error {
	if node.FileExists(path) {
		return errors.New("database already exists")
	}

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return err
	}

	db.AutoMigrate(
		&node.Head{},
		node.RootParms{},
	)

	slog.Info("State database created.", "path", path)
	return nil
}
