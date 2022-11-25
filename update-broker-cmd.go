package main

import (
	"flag"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func init() {
	commands["update-broker"] = newBrokerCommand
}

func updateBrokerCommand() {
	log.Debug("Running 'update-broker' command.")

	// var bro broker
	var (
		oldName     string
		udName      string
		udAccount   string
		udStatus    string
		udBase      float64
		udQuote     float64
		udMinWait   float64
		udMaxWait   float64
		udHighLimit float64
		udLowLimit  float64
		udDelta     float64
		udOffset    float64
	)

	flags := flag.NewFlagSet("new-broker", flag.ExitOnError)
	flags.StringVar(&oldName, "n", "defaultBroker", "Name of the broker to update.")
	flags.StringVar(&udName, "N", "defaultBroker", "New name of the broker.")
	flags.StringVar(&udAccount, "a", "defaultAccount", "Name of the account the updated broker will belong to.")
	flags.StringVar(&udStatus, "s", "disabled", "New status of the broker.")
	flags.Float64Var(&udBase, "b", 0, "New amount of 'base' currency handled by the broker.")
	flags.Float64Var(&udQuote, "q", 0, "New amount of 'quote' currency handled by the broker.")
	flags.Float64Var(&udMinWait, "w", 3600, "New minimum wait time between trades in seconds.")
	flags.Float64Var(&udMaxWait, "x", 10800, "New maximum wait time between trades in seconds.")
	flags.Float64Var(&udHighLimit, "t", 5, "New high limit.")
	flags.Float64Var(&udLowLimit, "l", 0.2, "New low limit.")
	flags.Float64Var(&udDelta, "d", 0.04, "New minimum price change between trades.")
	flags.Float64Var(&udOffset, "o", 0.0025, "New offset of limit trades from current price.")

	flags.Parse(flag.Args()[1:])

	db := openDB()
	defer db.Close()

	// tx, err := db.Begin()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// sqlStmt := "SELECT accountID FROM account WHERE accountName = $1 ;"
	// if err := tx.QueryRow(sqlStmt, accountName).Scan(&bro.account); err != nil {
	// 	log.Fatal("Couldn't get account ID:", err)
	// }

	// log.Debug("Account found: ", bro.account)

	// sqlStmt = `INSERT INTO broker (brokerName, account) VALUES ($1, $2);`
	// _, err = tx.Exec(sqlStmt, bro.name, bro.account)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// sqlStmt = "SELECT brokerID FROM broker WHERE brokerName = $1 ;"
	// if err := tx.QueryRow(sqlStmt, bro.name).Scan(&bro.id); err != nil {
	// 	log.Fatal("Couldn't get broker ID:", err)
	// }

	// sqlStmt = `
	// 	INSERT INTO brokerSetting
	// 		( broker, status, minWait, maxWait, highLimit, lowLimit, delta, offset )
	// 		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	// `
	// _, err = tx.Exec(sqlStmt,
	// 	bro.id, bro.status, bro.minWait, bro.maxWait, bro.highLimit, bro.lowLimit, bro.delta, bro.offset)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// sqlStmt = `INSERT INTO brokerBalance (broker, base, quote) VALUES ($1, $2, $3);`
	// _, err = tx.Exec(sqlStmt, bro.id, bro.base, bro.quote)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Debug("Broker inserted: ", bro.name, bro.id)

	// err = tx.Commit()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	log.Debug("Updated broker:")
}
