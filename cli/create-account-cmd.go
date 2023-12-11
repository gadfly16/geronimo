package cli

import (
	"flag"
	"fmt"

	mt "github.com/gadfly16/geronimo/messagetypes"
	"github.com/gadfly16/geronimo/server"

	log "github.com/sirupsen/logrus"
)

func init() {
	Commands["create-account"] = createAccountCLI
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
	acc.PasswordHash = server.HashPassword(password)

	if apiPublicKey == "" {
		apiPublicKey = getTerminalString(fmt.Sprintf("Enter API public key for account `%s`: ", acc.Name))
	}
	acc.ApiPublicKey = server.EncryptString(password, acc.Name, apiPublicKey)

	if apiPrivateKey == "" {
		apiPrivateKey = getTerminalString(fmt.Sprintf("Enter API private key for account `%s`: ", acc.Name))
	}
	acc.ApiPrivateKey = server.EncryptString(password, acc.Name, apiPrivateKey)

	// Connect to server
	conn, err := connectWSServer(s.WSAddr)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer server.CloseServerConnection(conn)

	msg := server.Message{
		Type:    mt.CreateAccount,
		Payload: acc,
	}

	err = msg.SendWSMessage(conn)
	if err != nil {
		log.Fatalln("Error during sending command: ", err)
	}

	resp, err := server.ReceiveWSMessage(conn)
	if err != nil {
		return err
	}

	if resp.Type == mt.Error {
		return resp.Payload.(error)
	}

	log.Info("Account created succesfully.")
	return nil
}
