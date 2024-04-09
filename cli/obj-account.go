package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gadfly16/geronimo/server"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

func init() {
	addObjectFlags(accountCmd)
	// accountCmd.PersistentFlags().UintVarP(&act.node.ParentID, "parent-id", "i", 0, "parent ID of the new account")
	accountCmd.PersistentFlags().StringVarP(&act.acc.APIPublicKey, "public-key", "k", "", "API public key")
	accountCmd.PersistentFlags().StringVarP(&act.acc.APIPrivateKey, "private-key", "K", "", "API private key")
	createCmd.AddCommand(accountCmd)
}

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "defines an account to operate on",
	Long: `The 'account' command defines the details of an account object
			so that different actions can be performed on them.`,
	Run: runAccount,
}

func runAccount(cmd *cobra.Command, args []string) {
	log.Debug("Executing 'new account' command.")

	act.node.DetailType = server.NodeAccount
	act.node.Detail = act.acc

	act.acc.Status = server.StatusKinds[act.status]

	if act.acc.APIPublicKey == "" {
		act.acc.APIPublicKey = getTerminalPassword(fmt.Sprintf("Enter API public key for account `%s`: ", act.node.Name))
	}

	if act.acc.APIPrivateKey == "" {
		act.acc.APIPrivateKey = getTerminalPassword(fmt.Sprintf("Enter API private key for account `%s`: ", act.node.Name))
	}

	conn, err := connectServer(&s)
	if err != nil {
		cliError(err)
		return
	}

	if act.node.ParentID == 0 {
		uid, err := strconv.Atoi(conn.claims.StandardClaims.Subject)
		if err != nil {
			cliError(err)
			return
		}
		act.node.ParentID = uint(uid)
	}

	resp, err := conn.client.R().
		SetBody(act.node).
		SetError(&server.APIError{}).
		Post("/api" + server.APIAccount)
	if err != nil {
		cliError(err)
		return
	}
	if resp.StatusCode() >= 400 {
		cliError(errors.New(resp.Error().(*server.APIError).Error))
		return
	}
}
