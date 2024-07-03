package cli

import (
	"errors"
	"log/slog"

	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gadfly16/geronimo/msg"
	"github.com/gadfly16/geronimo/node"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initializes database and secret keys",
	Long: `The 'init' command initializes all required files in the 
			working directory for the application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := InitDb(sdb); err != nil {
			slog.Error("Failed to create db. Exiting.", "error", err.Error())
			return
		}
		if err := node.ConnectDB(sdb); err != nil {
			slog.Error("Failed to connect to db. Exiting.", "error", err.Error())
			return
		}

		r := &node.RootNode{
			Head: &node.Head{
				Name: "Root",
				Kind: node.RootKind,
			},
			Parms: &rp,
		}
		if err := r.Init(); err != nil {
			slog.Error("Failed to create root node. Exiting.", "error", err.Error())
			return
		}
		resp := (&msg.Msg{
			Kind: msg.CreateKind,
			Payload: &node.GroupNode{
				Head: &node.Head{
					Name: "Users",
					Kind: node.GroupKind,
				},
			},
		}).Ask(node.Tree.Root)
		if resp.Kind == msg.ErrorKind {
			slog.Error("User group creation failed. Exiting!", "error", resp.Error())
		}
		(&msg.Msg{Kind: msg.StopRootKind}).Ask(node.Tree.Root)
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
