package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gadfly16/geronimo/node"
	"github.com/spf13/cobra"
)

var (
	rp           node.RootParms
	sdb          string
	userEmail    string
	userPassword string

	runtimeErr bool
)

func init() {
	// err := node.ConnectDB()
	// r := node.LoadRootParms()
	// if r != nil {
	// 	rp = r
	// }
	rootCmd.PersistentFlags().StringVarP(&rp.LogLevel, "log-level", "L", "debug", "logging level")
	rootCmd.PersistentFlags().StringVarP(&sdb, "state-db", "S", os.Getenv("HOME")+"/.config/Nerd/state.db", "state database path")
	rootCmd.PersistentFlags().StringVarP(&rp.HTTPAddr, "http-address", "A", "localhost:8088", "server HTTP address")
	rootCmd.PersistentFlags().StringVarP(&userEmail, "user-email", "U", "", "login email address")
	rootCmd.PersistentFlags().StringVarP(&userPassword, "user-password", "P", "", "login password")
}

var rootCmd = &cobra.Command{
	Use:   "geronimo",
	Short: "Geronimo is an crypto investment platform",
	Long:  `Geronimo is a web application to track, manage and automate crypto investments.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		slog.Debug("Inside root command's persistent pre run.")
	},
}

func Execute() {
	cobra.EnableTraverseRunHooks = true
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if runtimeErr {
		os.Exit(1)
	}
}
