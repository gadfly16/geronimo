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

	var (
		acc    server.AccountDetail
		secret server.AccountSecret
	)

	flags := flag.NewFlagSet("new-account", flag.ExitOnError)
	flags.StringVar(&acc.Name, "n", "defaultAccount", "Name of the new account.")
	flags.StringVar(&acc.Status, "s", "active", "Status of accout. ('active', 'disabled')")
	flags.StringVar(&secret.ApiPublicKey, "k", "", "API public key.")
	flags.StringVar(&secret.ApiPrivateKey, "K", "", "API private key.")
	flags.UintVar(&acc.UserID, "user-id", 0, "User ID of the owner of the account.")

	flags.Parse(flag.Args()[1:])

	if secret.ApiPublicKey == "" {
		secret.ApiPublicKey = getTerminalPassword(fmt.Sprintf("Enter API public key for account `%s`: ", acc.Name))
	}

	if secret.ApiPrivateKey == "" {
		secret.ApiPrivateKey = getTerminalPassword(fmt.Sprintf("Enter API private key for account `%s`: ", acc.Name))
	}

	conn, err := connectServer(&s)
	if err != nil {
		return err
	}

	if acc.UserID == 0 {
		uid, err := strconv.Atoi(conn.claims.StandardClaims.Subject)
		if err != nil {
			return err
		}
		acc.UserID = uint(uid)
	}

	resp, err := conn.client.R().
		SetBody(server.AccountWithSecret{Account: &acc, Secret: &secret}).
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
