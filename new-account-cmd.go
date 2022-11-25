package main

import (
	"flag"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func init() {
	commands["new-account"] = newAccountCommand
}

func newAccountCommand() {
	log.Debug("Running 'new-account' command.")

	var acc account

	naFlags := flag.NewFlagSet("new-account", flag.ExitOnError)
	naFlags.StringVar(&acc.name, "n", "defaultAccount", "Name of the new account.")
	naFlags.StringVar(&acc.publicKey, "k", "", "Public part of the API key.")
	naFlags.StringVar(&acc.privateKey, "p", "", "Private part of the API key.")

	naFlags.Parse(flag.Args()[1:])

	db := openDB()
	defer db.Close()

	sqlStmt := `INSERT INTO account (name, apiPublic, apiPrivate) VALUES ($1, $2, $3);`
	_, err := db.Exec(sqlStmt, acc.name, acc.publicKey, acc.privateKey)
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Created new account.")
}
