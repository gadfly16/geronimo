package cli

import (
	"errors"
	"flag"
	"fmt"

	mt "github.com/gadfly16/geronimo/messagetypes"
	"github.com/gadfly16/geronimo/server"

	log "github.com/sirupsen/logrus"
)

func init() {
	CommandHandlers["create-account"] = createAccountCLI
}

func createAccountCLI(s server.Settings) error {
	log.Debug("Running 'create-account' command.")

	var (
		acc           server.Account
		password      string
		apiPublicKey  string
		apiPrivateKey string
	)

	naFlags := flag.NewFlagSet("new-account", flag.ExitOnError)
	naFlags.StringVar(&acc.Name, "n", "defaultAccount", "Name of the new account.")
	naFlags.StringVar(&acc.Status, "s", "active", "Status of accout. ('active', 'disabled')")
	naFlags.StringVar(&password, "p", "", "Password of the new account.")
	naFlags.StringVar(&apiPublicKey, "u", "", "Public part of the API key.")
	naFlags.StringVar(&apiPrivateKey, "r", "", "Private part of the API key.")

	naFlags.Parse(flag.Args()[1:])

	if password == "" {
		password = getTerminalString(fmt.Sprintf("Enter password for account `%s`: ", acc.Name))
	}
	// acc.PasswordHash = server.HashPassword(password)

	if apiPublicKey == "" {
		apiPublicKey = getTerminalString(fmt.Sprintf("Enter API public key for account `%s`: ", acc.Name))
	}
	acc.ApiPublicKey = server.EncryptString(password, acc.Name, apiPublicKey)

	if apiPrivateKey == "" {
		apiPrivateKey = getTerminalString(fmt.Sprintf("Enter API private key for account `%s`: ", acc.Name))
	}
	acc.ApiPrivateKey = server.EncryptString(password, acc.Name, apiPrivateKey)

	// Connect to server
	conn, clientID, err := connectWSServer(s.WSAddr)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}

	msg := &server.Message{
		Type:     mt.CreateAccount,
		Payload:  acc,
		ID:       1,
		ClientID: clientID,
	}

	err = msg.SendWSMessage(conn)
	if err != nil {
		log.Fatalln("Error during sen ding command: ", err)
	}

	resp, err := waitForResponse(conn, msg)
	if err != nil {
		return err
	}
	if resp.Type == mt.Error {
		return resp.Payload.(error)
	}
	if resp.Type != mt.NewAccount {
		return errors.New(fmt.Sprint("not the expected response type: ", resp.Type))
	}

	log.Info("Account created succesfully: ", resp.Payload.(server.Account).Name)

	return closeServerConnection(conn)
}
