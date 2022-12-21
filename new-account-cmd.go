package main

import (
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"fmt"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
)

func init() {
	commands["new-account"] = newAccountCommand
}

func newAccountCommand() {
	log.Debug("Running 'new-account' command.")

	var acc account
	var rawPassword string

	naFlags := flag.NewFlagSet("new-account", flag.ExitOnError)
	naFlags.StringVar(&acc.name, "n", "defaultAccount", "Name of the new account.")
	naFlags.StringVar(&rawPassword, "p", "", "Password of the new account.")
	naFlags.StringVar(&acc.apiPublicKey, "u", "", "Public part of the API key.")
	naFlags.StringVar(&acc.apiPrivateKey, "r", "", "Private part of the API key.")

	naFlags.Parse(flag.Args()[1:])

	if rawPassword == "" {
		fmt.Printf("Enter password for account `%s`: ", acc.name)
		pwA, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal("Couldn't get password.")
		}
		fmt.Printf("\nPlease re-enter password for account `%s`: ", acc.name)
		pwB, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal("Couldn't get password.")
		}
		if string(pwA) != string(pwB) {
			log.Fatal("Passwords don't match.")
		}
		rawPassword = string(pwA)
	}

	h := sha256.New()
	h.Write([]byte(rawPassword))
	acc.password = base64.StdEncoding.EncodeToString(h.Sum(nil))

	db := openDB()
	defer db.Close()

	sqlStmt := `INSERT INTO account (name, password, apiPublicKey, apiPrivateKey) VALUES ($1, $2, $3, $4);`
	_, err := db.Exec(sqlStmt, acc.name, acc.password, acc.apiPublicKey, acc.apiPrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Created new account.")
}
