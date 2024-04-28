package cli

import (
	"strconv"
	"strings"

	"github.com/gadfly16/geronimo/server"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var holdingStrings []string

func init() {
	addObjectFlags(pocketCmd)
	pocketCmd.Flags().StringSliceVarP(&holdingStrings, "holding", "H", []string{}, "comma separated list of asset:value pairs that comprise the holdings in the pocket")
	createCmd.AddCommand(pocketCmd)
}

var pocketCmd = &cobra.Command{
	Use:   "pocket",
	Short: "defines pocket to operate on",
	Long: `The 'pocket' command defines the details of a pocket object
			so that different actions can be performed on them.`,
	Run: runPocket,
}

func runPocket(cmd *cobra.Command, args []string) {
	log.Debugln("Executing", act.cmd, "on group object")

	act.group.Status = server.StatusKinds[act.status]

	for _, hs := range holdingStrings {
		parts := strings.Split(hs, ":")
		if len(parts) != 2 {
			cliError("holdings must be given in 'sym:value' format")
			return
		}
		ass, ok := server.Assets[strings.ToLower(parts[0])]
		if !ok {
			cliError("Asset " + parts[0] + "not supported.")
			return
		}
		val, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			cliError(err.Error())
			return
		}
		act.pocket.Holdings[ass] = val
	}

	act.node.DetailType = server.NodePocket
	act.node.Detail = &act.pocket
	act.msg.Payload = &act.node

	conn, err := connectServer(&s)
	if err != nil {
		cliError(err.Error())
		return
	}

	route := "/api" + server.APICreate + "/pocket"
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
