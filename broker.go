package main

import (
	"database/sql"

	log "github.com/sirupsen/logrus"
)

type broker struct {
	id        int64
	name      string
	accountId int64
	status    string
	minWait   float64
	maxWait   float64
	highLimit float64
	lowLimit  float64
	delta     float64
	offset    float64
	base      float64
	quote     float64
}

// func getBroker(name string) *broker {
// 	db := openDB()
// 	defer db.Close()

// 	sqlStmt = `
// 		SELECT * FROM broker b
// 		JOIN brokerSetting bs ON b.id=bs.broker
// 		JOIN brokerBalance bb ON b.id=bb.broker ;
// 	`
// }

func newBrokerSetting(tx *sql.Tx, bro *broker) {
	sqlStmt := `
		INSERT INTO brokerSetting
			(brokerId, status, minWait, maxWait, highLimit, lowLimit, delta, offset)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := tx.Exec(sqlStmt,
		bro.id, bro.status, bro.minWait, bro.maxWait, bro.highLimit, bro.lowLimit, bro.delta, bro.offset)
	if err != nil {
		log.Fatal(err)
	}
}

func newBrokerBalance(tx *sql.Tx, bro *broker) {
	sqlStmt := `INSERT INTO brokerBalance (brokerId, base, quote) VALUES ($1, $2, $3)`
	_, err := tx.Exec(sqlStmt, bro.id, bro.base, bro.quote)
	if err != nil {
		log.Fatal(err)
	}
}

func runBroker(bro broker) {
	log.Debug("Running broker: ", bro.name)
}
