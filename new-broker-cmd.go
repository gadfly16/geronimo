package main

import (
	"flag"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func init() {
	commands["new-broker"] = newBrokerCommand
}

func newBrokerCommand() {
	log.Debug("Running 'new-broker' command.")

	var bro broker
	var accountName string

	flags := flag.NewFlagSet("new-broker", flag.ExitOnError)
	flags.StringVar(&bro.name, "n", "defaultBroker", "Name of the new broker.")
	flags.StringVar(&accountName, "a", "defaultAccount", "Name of the account the new broker belongs to.")
	flags.StringVar(&bro.pair, "p", "ADA/EUR", "The asset pair's name the broker trades with.")
	flags.StringVar(&bro.status, "s", "disabled", "Status of the new broker.")
	flags.Float64Var(&bro.base, "b", 0, "Amount of 'base' currency handled by the broker.")
	flags.Float64Var(&bro.quote, "q", 0, "Amount of 'quote' currency handled by the broker.")
	flags.Float64Var(&bro.minWait, "w", 3600, "Minimum wait time between trades in seconds.")
	flags.Float64Var(&bro.maxWait, "x", 10800, "Maximum wait time between trades in seconds.")
	flags.Float64Var(&bro.highLimit, "t", 5, "High limit.")
	flags.Float64Var(&bro.lowLimit, "l", 0.2, "Low limit.")
	flags.Float64Var(&bro.delta, "d", 0.04, "Minimum price change between trades.")
	flags.Float64Var(&bro.offset, "o", 0.0025, "Offset of limit trades from current price.")

	flags.Parse(flag.Args()[1:])

	db := openDB()
	defer db.Close()

	bro.accountId = getAccountID(db, accountName)

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	sqlStmt := `INSERT INTO brokerHead (name, accountId, pair)
		VALUES ($1, $2, $3);`
	_, err = tx.Exec(sqlStmt, bro.name, bro.accountId, bro.pair)
	if err != nil {
		log.Fatal(err)
	}

	sqlStmt = "SELECT id FROM brokerHead WHERE name = $1 ;"
	if err := tx.QueryRow(sqlStmt, bro.name).Scan(&bro.id); err != nil {
		log.Fatal("Couldn't get broker ID:", err)
	}

	newBrokerSetting(tx, &bro)
	newBrokerBalance(tx, &bro)
	log.Debug("Broker inserted: ", bro.name, bro.id)

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Created new broker.")
}
