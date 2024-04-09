package cli

import (
	"fmt"
	"os"

	"github.com/gadfly16/geronimo/server"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	s            server.Settings
	userEmail    string
	userPassword string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&s.LogLevel, "log-level", "L", "debug", "logging level")
	rootCmd.PersistentFlags().StringVarP(&s.WorkDir, "work-dir", "W", os.Getenv("HOME")+"/.config/Geronimo", "work dirextory for persistent storage")
	rootCmd.PersistentFlags().StringVarP(&s.HTTPAddr, "http-address", "A", "localhost:8088", "server HTTP address")
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
		s.Init()
		log.Debugln("Inside root command's persistent pre run.")
		fmt.Println(s.DBPath)
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
