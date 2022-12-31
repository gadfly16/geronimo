package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func init() {
	commands["run"] = runCommand
}

func runCommand() {
	log.Debug("Running 'run' command.")

	flags := flag.NewFlagSet("run", flag.ExitOnError)
	flags.Parse(flag.Args()[1:])

	db := openDB()
	defer db.Close()

	sqlStmt := `
		SELECT DISTINCT a.id, a.name, a.pwhash, a.apiPublicKey, a.apiPrivateKey
		FROM account a
		JOIN broker b WHERE b.accountId = a.id
		AND b.status = "active" `
	rows, err := db.Query(sqlStmt)
	if err != nil {
		log.Fatal("Couldn't get accounts with active brokers: ", err)
	}
	defer rows.Close()

	var activeAccs []account

	for rows.Next() {
		var acc account
		if err := rows.Scan(&acc.id, &acc.name, &acc.pwhash, &acc.apiPublicKey, &acc.apiPrivateKey); err != nil {
			log.Fatal("Couldn't create account: ", err)
		}
		activeAccs = append(activeAccs, acc)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	for _, acc := range activeAccs {
		decryptAccountKeys(&acc)
		go runBookkeeper(acc)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	for range signals {
		log.Warn("Stopping...")
		return
	}
}
