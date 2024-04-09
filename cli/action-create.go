package cli

import (
	"github.com/gadfly16/geronimo/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	addActionFlags(createCmd)
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "creates new objects",
	Long:  `The new command creates all kind of objects on the server.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.Debugln("Inside create command's pre-run.")
		act.msg.Type = server.MessageCreate
	},
}
