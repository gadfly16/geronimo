package main

import (
	"database/sql"
	"math"
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

	lastOrd := getLastOrder(db, bro)
	if lastOrd == nil {
		log.Debug("No previous order found, doing immediate trade.")
	} else {
		elapsed := time.Since(lastOrd.tstamp)
		wait := time.Duration(fit01(rand.Float64(), bro.minWait, bro.maxWait))*time.Second - elapsed
		log.Debugf("Waiting for next check: %s", wait.String())
		time.Sleep(wait)
	}

	ord := order{brokerId: bro.id}
	log.Debugf("Broker `%s` is asking for `midPrice`.", bro.name)
	orders <- ord
	pricedOrd := <-receipt
	log.Debugf("Broker `%s` received `midPrice`: %v",
		bro.name, pricedOrd.midPrice)
	pricedOrd.prepareTrade(&bro, lastOrd)
	if pricedOrd.price == 0 {
		log.Info("No order necessary.")
	} else {
		log.Infof("Requesting order placement by `%s`: %v @ %v", bro.name, pricedOrd.volume, pricedOrd.price)
		orders <- pricedOrd
	}
}

func (ord *order) prepareTrade(bro *broker, lastOrd *order) {
	if lastOrd == nil ||
		math.Abs(ord.midPrice-lastOrd.price)/lastOrd.price > bro.delta ||
		ord.midPrice > bro.highLimit || ord.midPrice < bro.lowLimit {
		diff := getVolume(ord.midPrice, bro.lowLimit, bro.highLimit, bro.base, bro.quote)
		if diff > 0 {
			ord.price = ord.midPrice / (1 + bro.offset)
		} else {
			ord.price = ord.midPrice * (1 + bro.offset)
		}
		ord.volume = getVolume(ord.price, bro.lowLimit, bro.highLimit, bro.base, bro.quote)
	}
}

func getBias(price, low, high float64) float64 {
	return clamp01(fitTo01(math.Log(price), math.Log(high), math.Log(low)))
}

func getVolume(price, low, high, base, quote float64) float64 {
	bias := getBias(price, low, high)
	baseValue := base * price
	allValue := baseValue + quote
	newBaseValue := allValue * bias
	return (newBaseValue - baseValue) / price
}
