package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/gadfly16/geronimo/server"
	"github.com/spf13/cobra"
)

func init() {
	treeCmd.Flags().UintVarP(&act.node.ID, "node-id", "i", 0, "node id")
	rootCmd.AddCommand(treeCmd)
}

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "gets the object tree from the server",
	Long:  `The tree command gets the tree structure from the server.`,
	Run:   runTree,
}

func runTree(cmd *cobra.Command, args []string) {
	conn, err := connectServer(&s)
	if err != nil {
		cliError(err)
		return
	}

	if act.node.ID == 0 {
		uid, err := strconv.Atoi(conn.claims.StandardClaims.Subject)
		if err != nil {
			cliError(err)
			return
		}
		act.node.ID = uint(uid)
	}

	resp, err := conn.client.R().
		SetError(&server.APIError{}).
		SetQueryParam("userid", strconv.Itoa(int(act.node.ID))).
		Get("/api" + server.APITree)
	if err != nil {
		cliError(err)
		return
	}
	if resp.StatusCode() >= 400 {
		cliError(errors.New(resp.Error().(*server.APIError).Error))
		return
	}

	state := map[string]any{}
	if err = json.Unmarshal(resp.Body(), &state); err != nil {
		cliError(err)
		return
	}
	output, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		cliError(err)
		return
	}
	fmt.Println(string(output))
}
