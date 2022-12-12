package main

import (
	"database/sql"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

type broker struct {
	id        int64
	name      string
	pair      string
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

func runBroker(bro broker, orders, receipt chan order) {
	log.Debug("Running broker: ", bro.name)

	db := openDB()
	defer db.Close()

	lastOrd := lastOrder(db, bro)
	if lastOrd == nil {
		log.Debug("No previous order found, doing immediate trade.")
	} else {
		elapsed := time.Since(lastOrd.tstamp)
		wait := time.Duration(fit01(rand.Float64(), bro.minWait, bro.maxWait))*time.Second - elapsed
		log.Debug("Waiting for next check: %s", wait.String())
		time.Sleep(wait)
	}

	ord := order{brokerId: bro.id}
	log.Debugf("Broker `%s` is asking for ticker.", bro.name)
	orders <- ord
	log.Debugf("Broker `%s` is waiting for ticker.", bro.name)
	res := <-receipt
	log.Debugf("Broker `%s` received ticker. Ask: %v Bid: %v",
		bro.name, res.ticker.Ask.Price, res.ticker.Bid.Price)
}
