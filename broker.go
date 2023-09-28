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
	acc *account

	id        int64
	name      string
	pair      string
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

	msg      chan brokerMsg
	receipts chan order
}

func (bro *broker) newSetting(tx *sql.Tx) {
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

func (bro *broker) newBalance(tx *sql.Tx) {
	sqlStmt := `INSERT INTO brokerBalance (brokerId, base, quote, fee) VALUES ($1, $2, $3, $4)`
	_, err := tx.Exec(sqlStmt, bro.id, bro.base, bro.quote, bro.fee)
	if err != nil {
		log.Fatal(err)
	}
}

func (bro *broker) getLastOrder(db *sql.DB) *order {
	o := &order{}
	var ts int64
	var brokerId int64
	sqlStmt := `SELECT * FROM 'order'
		WHERE brokerId = $1 AND completed > 0
		ORDER BY tstamp DESC LIMIT 1`
	if err := db.QueryRow(sqlStmt, bro.id).Scan(&o.userRef, &brokerId, &o.status,
		&o.orderId, &o.volume, &o.completed, &o.price, &ts); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Fatal("Couldn't get last order: ", err)
	}
	o.bro = bro
	o.tstamp = time.UnixMilli(ts)
	return o
}

func (bro *broker) run() {
	log.Debug("Running broker: ", bro.name)

	db := openDB()
	defer db.Close()
	var lastCheck time.Time

	lastOrd := bro.getLastOrder(db)
	if lastOrd == nil {
		log.Debug("No previous order found, last check set to epic.")
	} else {
		lastCheck = lastOrd.tstamp
	}

	for {
		log.Debugf("Cycle started for broker: %+v", bro)
		elapsed := time.Since(lastCheck)
		wait := time.Duration(fit01(rand.Float64(), bro.minWait, bro.maxWait))*time.Second - elapsed
		if wait < 0 {
			log.Debug("Immediate check needed.")
		} else {
			log.Debugf("Waiting for next check: %v", wait)
			time.Sleep(wait)
		}

		// Ask for price
		ord := order{bro: bro}
		bro.acc.orders <- ord
		pricedOrd := <-bro.receipts

		pricedOrd.fillOrder()

		if pricedOrd.price == 0 {
			log.Info("No order necessary.")
		} else if math.Abs(pricedOrd.volume) < krakenMinTradeVolume(bro.pair) {
			log.Infof("Volume '%v' smaller than kraken min volume for pair: %v (%v)",
				pricedOrd.volume, bro.pair, krakenMinTradeVolume(bro.pair))
		} else {
			log.Infof("Requesting order placement by `%s`: %v @ %v", bro.name, pricedOrd.volume, pricedOrd.price)
			bro.acc.orders <- pricedOrd
		}
		lastCheck = time.Now()
	}
}

func getBias(price, low, high float64) float64 {
	return clamp01(fitTo01(math.Log(price), math.Log(high), math.Log(low)))
}

// getVolume calculates the volume of the order to be made
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
	bro.newBalance(tx)

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
