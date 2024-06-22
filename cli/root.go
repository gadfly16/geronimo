package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gadfly16/geronimo/node/root"
	"github.com/spf13/cobra"
)

var (
	rp           root.RootParms
	userEmail    string
	userPassword string

	runtimeErr bool
)

func init() {
	// rootCmd.PersistentFlags().StringVarP(&rp.LogLevel, "log-level", "L", "debug", "logging level")
	rootCmd.PersistentFlags().StringVarP(&rp.WorkDir, "work-dir", "W", os.Getenv("HOME")+"/.config/Nerd", "work dirextory for persistent storage")
	rootCmd.PersistentFlags().StringVarP(&rp.HTTPAddr, "http-address", "A", "localhost:8088", "server HTTP address")
	rootCmd.PersistentFlags().StringVarP(&userEmail, "user-email", "U", "", "login email address")
	rootCmd.PersistentFlags().StringVarP(&userPassword, "user-password", "P", "", "login password")
}

var rootCmd = &cobra.Command{
	Use:   "geronimo",
	Short: "Geronimo is an crypto investment platform",
	Long:  `Geronimo is a web application to track, manage and automate crypto investments.`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	log.Debugln("Inside root command.")
	// },
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		rp.DbPath = rp.WorkDir + "/state.db"
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
