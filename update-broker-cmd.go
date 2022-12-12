package main

import (
	"flag"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func init() {
	commands["update-broker"] = updateBrokerCommand
}

func updateBrokerCommand() {
	log.Debug("Running 'update-broker' command.")

	// var bro broker
	var (
		oldName       string
		udName        string
		udAccountName string
		udStatus      string
		udBase        float64
		udQuote       float64
		udMinWait     float64
		udMaxWait     float64
		udHighLimit   float64
		udLowLimit    float64
		udDelta       float64
		udOffset      float64
	)

	flags := flag.NewFlagSet("new-broker", flag.ExitOnError)
	flags.StringVar(&oldName, "n", "defaultBroker", "Name of the broker to update.")
	flags.StringVar(&udName, "N", "", "New name of the broker.")
	flags.StringVar(&udAccountName, "a", "", "Name of the account the updated broker will belong to.")
	flags.StringVar(&udStatus, "s", "", "New status of the broker.")
	flags.Float64Var(&udBase, "b", -1, "New amount of 'base' currency handled by the broker.")
	flags.Float64Var(&udQuote, "q", -1, "New amount of 'quote' currency handled by the broker.")
	flags.Float64Var(&udMinWait, "w", -1, "New minimum wait time between trades in seconds.")
	flags.Float64Var(&udMaxWait, "x", -1, "New maximum wait time between trades in seconds.")
	flags.Float64Var(&udHighLimit, "t", -1, "New high limit.")
	flags.Float64Var(&udLowLimit, "l", -1, "New low limit.")
	flags.Float64Var(&udDelta, "d", -1, "New minimum price change between trades.")
	flags.Float64Var(&udOffset, "o", -1, "New offset of limit trades from current price.")

	flags.Parse(flag.Args()[1:])

	db := openDB()
	defer db.Close()

	var bro broker

	log.Debug("Getting broker for update: ", oldName)

	sqlStmt := `SELECT * FROM broker WHERE name = $1`
	if err := db.QueryRow(sqlStmt, oldName).Scan(&bro.id, &bro.accountId, &bro.name,
		&bro.pair, &bro.status, &bro.minWait, &bro.maxWait, &bro.highLimit, &bro.lowLimit,
		&bro.delta, &bro.offset, &bro.base, &bro.quote); err != nil {
		log.Fatal("Couldn't get broker for update: ", err)
	}

	log.Debug("Got broker for update: ", bro)

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	change := false
	if udName != "" {
		log.Debug("Changing broker name to: ", udName)
		bro.name = udName
		change = true
	}
	if udAccountName != "" {
		log.Debug("Changing account to: ", udAccountName)
		bro.accountId = getAccountID(db, udAccountName)
		change = true
	}
	if change {
		sqlStmt = `UPDATE brokerHead SET name=$1, accountId=$2 WHERE id=$3`
		tx.Exec(sqlStmt, bro.name, bro.accountId, bro.id)
	}

	change = false
	if udStatus != "" {
		log.Debug("Changing broker status to: ", udStatus)
		bro.status = udStatus
		change = true
	}
	if udMinWait != -1 {
		log.Debug("Changing minWait to: ", udMinWait)
		bro.minWait = udMinWait
		change = true
	}
	if udMaxWait != -1 {
		log.Debug("Changing maxWait to: ", udMaxWait)
		bro.maxWait = udMaxWait
		change = true
	}
	if udHighLimit != -1 {
		log.Debug("Changing highLimit to: ", udHighLimit)
		bro.highLimit = udHighLimit
		change = true
	}
	if udLowLimit != -1 {
		log.Debug("Changing lowLimit to: ", udLowLimit)
		bro.lowLimit = udLowLimit
		change = true
	}
	if udDelta != -1 {
		log.Debug("Changing delta to: ", udDelta)
		bro.delta = udDelta
		change = true
	}
	if udOffset != -1 {
		log.Debug("Changing offset to: ", udOffset)
		bro.offset = udOffset
		change = true
	}
	if change {
		newBrokerSetting(tx, &bro)
	}

	change = false
	if udBase != -1 {
		log.Debug("Changing base to: ", udBase)
		bro.base = udBase
		change = true
	}
	if udQuote != -1 {
		log.Debug("Changing quote to: ", udQuote)
		bro.quote = udQuote
		change = true
	}
	if change {
		newBrokerBalance(tx, &bro)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Updated broker: ", bro.name)
}
