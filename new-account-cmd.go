package main

import (
	"flag"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func init() {
	commands["new-account"] = newAccountCommand
}

func newAccountCommand() {
	log.Debug("Running 'new-account' command.")

	var (
		acc           account
		password      string
		apiPublicKey  string
		apiPrivateKey string
	)

	naFlags := flag.NewFlagSet("new-account", flag.ExitOnError)
	naFlags.StringVar(&acc.name, "n", "defaultAccount", "Name of the new account.")
	naFlags.StringVar(&password, "p", "", "Password of the new account.")
	naFlags.StringVar(&apiPublicKey, "u", "", "Public part of the API key.")
	naFlags.StringVar(&apiPrivateKey, "r", "", "Private part of the API key.")

	naFlags.Parse(flag.Args()[1:])

	if password == "" {
		password = getTerminalString(fmt.Sprintf("Enter password for account `%s`: ", acc.name))
	}
	acc.pwhash = hashPassword(password)

	if apiPublicKey == "" {
		apiPublicKey = getTerminalString(fmt.Sprintf("Enter API public key for account `%s`: ", acc.name))
	}
	acc.apiPublicKey = encryptString(password, acc.name, apiPublicKey)

	if apiPrivateKey == "" {
		apiPrivateKey = getTerminalString(fmt.Sprintf("Enter API public key for account `%s`: ", acc.name))
	}
	acc.apiPrivateKey = encryptString(password, acc.name, apiPrivateKey)

	db := openDB()
	defer db.Close()

	sqlStmt := `INSERT INTO account (name, pwhash, apiPublicKey, apiPrivateKey) VALUES ($1, $2, $3, $4);`
	_, err := db.Exec(sqlStmt, acc.name, acc.pwhash, acc.apiPublicKey, acc.apiPrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Created new account.")
}
