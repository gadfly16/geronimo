package cli

import (
	"github.com/gadfly16/geronimo/server"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

func init() {
	addObjectFlags(groupCmd)
	createCmd.AddCommand(accountCmd)
}

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "defines group to operate on",
	Long: `The 'group' command defines the details of a group object
			so that different actions can be performed on them.`,
	Run: runGroup,
}

func runGroup(cmd *cobra.Command, args []string) {
	log.Debugln("Executing", act.cmd, "on group object")

	act.group.Status = server.StatusKinds[act.status]
	act.node.DetailType = server.NodeGroup
	act.node.Detail = &act.group
	act.msg.Payload = &act.node

	conn, err := connectServer(&s)
	if err != nil {
		cliError(err.Error())
		return
	}

	route := "/api" + server.APICreate + "/group"
	resp, err := conn.client.R().
		SetBody(&act.msg).
		SetError(&server.APIError{}).
		Post(route)
	if err != nil {
		cliError(err.Error())
		return
	}
	if resp.StatusCode() >= 400 {
		cliError(resp.Error().(*server.APIError).Error)
		return
	}
}
