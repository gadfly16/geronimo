package cli

import (
	"errors"
	"flag"
	"fmt"
	"strconv"

	"github.com/gadfly16/geronimo/server"

	log "github.com/sirupsen/logrus"
)

func init() {
	CommandHandlers["create-account"] = createAccountCLI
}

func createAccountCLI(s server.Settings) error {
	log.Debug("Running 'create-account' command.")

	acc := &server.Account{}
	node := &server.Node{
		DetailType: server.NodeAccount,
		Detail:     acc,
	}

	var status string

	flags := flag.NewFlagSet("new-account", flag.ExitOnError)
	flags.StringVar(&node.Name, "n", "defaultAccount", "Name of the new account.")
	flags.StringVar(&status, "s", "active", "Status of accout. ('active', 'disabled')")
	flags.StringVar(&acc.ApiPublicKey, "k", "", "API public key.")
	flags.StringVar(&acc.ApiPrivateKey, "K", "", "API private key.")
	flags.UintVar(&node.ParentID, "user-id", 0, "User ID of the owner of the account.")

	flags.Parse(flag.Args()[1:])

	var ok bool
	if acc.Status, ok = server.StatusKinds[status]; !ok {
		return errors.New("ivalid status kind")
	}

	if acc.ApiPublicKey == "" {
		acc.ApiPublicKey = getTerminalPassword(fmt.Sprintf("Enter API public key for account `%s`: ", node.Name))
	}

	if acc.ApiPrivateKey == "" {
		acc.ApiPrivateKey = getTerminalPassword(fmt.Sprintf("Enter API private key for account `%s`: ", node.Name))
	}

	conn, err := connectServer(&s)
	if err != nil {
		return err
	}

	if node.ParentID == 0 {
		uid, err := strconv.Atoi(conn.claims.StandardClaims.Subject)
		if err != nil {
			return err
		}
		node.ParentID = uint(uid)
	}

	resp, err := conn.client.R().
		SetBody(node).
		SetError(&server.APIError{}).
		Post("/api" + server.APIAccount)
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 400 {
		return errors.New(resp.Error().(*server.APIError).Error)
	}

	return nil
}
