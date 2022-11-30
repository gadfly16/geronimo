package main

import (
	"database/sql"

	log "github.com/sirupsen/logrus"
)

type account struct {
	id            int64
	name          string
	apiPublicKey  string
	apiPrivateKey string
}

func getAccountID(db *sql.DB, name string) int64 {
	var accountId int64
	sqlStmt := "SELECT id FROM account WHERE name = $1"
	if err := db.QueryRow(sqlStmt, name).Scan(&accountId); err != nil {
		log.Fatal("Couldn't get account ID:", err)
	}
	log.Debug("Account found: ", accountId)
	return accountId
}

func runBookkeeper(a account) {
	log.Debug("Running bookkeeper for: ", a.name)

	db := openDB()
	defer db.Close()

	sqlStmt := `
		SELECT * FROM broker
		WHERE accountId = $1
		AND status = "active"`
	rows, err := db.Query(sqlStmt, a.id)
	if err != nil {
		log.Fatal("Couldn't get brokers for account: ", err)
	}
	defer rows.Close()

	var bros []broker

	for rows.Next() {
		var bro broker
		if err := rows.Scan(&bro.id, &bro.name, &bro.accountId,
			&bro.status, &bro.minWait, &bro.maxWait, &bro.highLimit, &bro.lowLimit,
			&bro.delta, &bro.offset, &bro.base, &bro.quote); err != nil {
			log.Fatal("Couldn't scan broker for running: ", err)
		}
		bros = append(bros, bro)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	for _, bro := range bros {
		go runBroker(bro)
	}

}
