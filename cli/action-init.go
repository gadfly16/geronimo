package cli

import (
	"github.com/gadfly16/geronimo/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
		log.Debug("Running 'init' command.")
		if err := server.Init(s); err != nil {
			log.Error(err.Error())
			runtimeErr = true
		}
	},
}
