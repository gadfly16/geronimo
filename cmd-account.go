package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gadfly16/geronimo/server"

	log "github.com/sirupsen/logrus"
)

func init() {
	parser.AddCommand(
		"account",
		"account operations",
		"command that initiates operations on accounts",
		&accOpts)
}

var accOpts accountOptions

type accountOptions struct {
	Name          string `short:"n" long:"name" description:"name of the account" default:"NewAccount"`
	Status        string `short:"s" long:"status" default:"active" description:"status of the account" choice:"active" choice:"disabled"`
	APIPublicKey  string `short:"k" long:"public-key" description:"API public key"`
	APIPrivateKey string `short:"K" long:"private-key" description:"API private key"`
	ParentID      uint   `short:"u" long:"user-id" description:"Parent User ID"`

	NewCmd newAccountCommand `command:"new" description:"create a new account"`
}

type newAccountCommand struct{}

func (opts *newAccountCommand) Execute(args []string) error {
	s.Init()
	log.Debug("Executing 'new account' command.")

	acc := &server.Account{}
	node := &server.Node{
		DetailType: server.NodeAccount,
		Detail:     acc,
	}

	acc.Status = server.StatusKinds[accOpts.Status]

	if accOpts.APIPublicKey == "" {
		accOpts.APIPublicKey = getTerminalPassword(fmt.Sprintf("Enter API public key for account `%s`: ", node.Name))
	}
	acc.APIPublicKey = accOpts.APIPublicKey

	if accOpts.APIPrivateKey == "" {
		accOpts.APIPrivateKey = getTerminalPassword(fmt.Sprintf("Enter API private key for account `%s`: ", node.Name))
	}
	acc.APIPrivateKey = accOpts.APIPrivateKey

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
