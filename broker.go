package main

import (
	"database/sql"
	"math"
	"math/rand"
	"time"

	kws "github.com/aopoltorzhicky/go_kraken/websocket"
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
	fee       float64
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
	sqlStmt := `INSERT INTO brokerBalance (brokerId, base, quote, fee) VALUES ($1, $2, $3, $4)`
	_, err := tx.Exec(sqlStmt, bro.id, bro.base, bro.quote, bro.fee)
	if err != nil {
		log.Fatal(err)
	}
}

func runBroker(bro *broker, orders, receipt chan order) {
	log.Debug("Running broker: ", bro.name)

	db := openDB()
	defer db.Close()
	var lastCheck time.Time

	lastOrd := getLastOrder(db, bro)
	if lastOrd == nil {
		log.Debug("No previous order found, last check set to epic.")
	} else {
		lastCheck = lastOrd.tstamp
	}

	for {
		log.Debugf("Cycle started for broker: %+v", *bro)
		lastOrd = getLastOrder(db, bro)
		elapsed := time.Since(lastCheck)
		wait := time.Duration(fit01(rand.Float64(), bro.minWait, bro.maxWait))*time.Second - elapsed
		if wait < 0 {
			log.Debug("Immediate check needed.")
		} else {
			log.Debugf("Waiting for next check: %v", wait)
			time.Sleep(wait)
		}

		ord := order{brokerId: bro.id}
		orders <- ord
		pricedOrd := <-receipt
		pricedOrd.prepareTrade(bro, lastOrd)
		if pricedOrd.price == 0 {
			log.Info("No order necessary.")
		} else if pricedOrd.price < krakenMinTradeVolume(bro.pair) {
			log.Info("Volume '%v' smaller than kraken min volume for pair: %v", pricedOrd.price, bro.pair)
		} else {
			log.Infof("Requesting order placement by `%s`: %v @ %v", bro.name, pricedOrd.volume, pricedOrd.price)
			orders <- pricedOrd
		}
		lastCheck = time.Now()
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

func (bro *broker) bookTrade(db *sql.DB, tid string, ownTrd kws.OwnTrade, trdOrd *order) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	cost := jsonNumToFloat64(ownTrd.Cost)
	volume := jsonNumToFloat64(ownTrd.Vol)
	if ownTrd.Type == "sell" {
		volume *= -1
	} else {
		cost *= -1
	}

	trd := trade{
		id:      tid,
		orderId: ownTrd.OrderID,
		volume:  volume,
		cost:    cost,
		fee:     jsonNumToFloat64(ownTrd.Fee),
		tstamp:  time.UnixMilli(int64(jsonNumToFloat64(ownTrd.Time) * 1000)),
	}

	sqlStmt := `INSERT INTO trade VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = tx.Exec(sqlStmt, trd.id, trd.orderId, trd.volume,
		trd.cost, trd.fee, trd.tstamp.UnixMilli())
	if err != nil {
		log.Fatalf("Couldn't insert new trade '%v': %v", tid, err)
	}

	bro.base += trd.volume
	bro.quote += trd.cost
	bro.fee += trd.fee
	newBrokerBalance(tx, bro)

	sqlStmt = `UPDATE 'order' SET completed=completed+$1 WHERE orderId=$2`
	_, err = tx.Exec(sqlStmt, math.Abs(trd.volume), trd.orderId)
	if err != nil {
		log.Fatalf("Couldn't update order's completed value: %v", trd.orderId)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}
