package cli

import (
	"errors"
	"log/slog"

	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gadfly16/geronimo/node"
	"github.com/gadfly16/geronimo/node/all"
	"github.com/gadfly16/geronimo/node/root"
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
		if err := InitDb(rp.DbPath); err != nil {
			slog.Error("Failed to create db. Exiting.", "error", err.Error())
			return
		}
		if err := node.ConnectDB(rp.DbPath); err != nil {
			slog.Error("Failed to connect to db. Exiting.", "error", err.Error())
			return
		}

		r := &root.RootNode{
			Head: node.Head{
				Name: "Root",
				Kind: node.RootKind,
			},
			Parms: rp,
		}
		if err := r.Create(); err != nil {
			slog.Error("Failed to create root node. Exiting.", "error", err.Error())
			return
		}
		slog.Info("Geronimo initialized.")
	},
}

func InitDb(path string) error {
	if node.FileExists(path) {
		return errors.New("database already exists")
	}

	slog.Info("Creating settings database.", "path", path)

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return err
	}

	db.AutoMigrate(
		&node.Head{},
		all.RootParms,
	)

	slog.Info("database created", "path", path)
	return nil
}
