package main

import (
	"database/sql"
	"math"
	"time"

	kws "github.com/aopoltorzhicky/go_kraken/websocket"
	log "github.com/sirupsen/logrus"
)

type order struct {
	bro       *broker
	userRef   int64
	status    string
	orderId   string
	volume    float64
	completed float64
	price     float64
	midPrice  float64
	tstamp    time.Time
}

func (ord *order) fillOrder() {
	bro := ord.bro
	diff := getVolume(ord.midPrice, bro.lowLimit, bro.highLimit, bro.base, bro.quote)
	whole := bro.base + bro.quote/ord.midPrice
	log.Infof("Blance needs: %v", diff)
	inbalance := diff / whole
	if math.Abs(inbalance) < bro.delta {
		log.Infof("Inbalance under threshold: %.3f%% (%.3f%%)", inbalance*100, bro.delta)
		return
	}
	log.Infof("Inbalance requires order placement: %.3f%%", inbalance*100)
	if diff > 0 {
		ord.price = ord.midPrice / (1 + bro.offset)
	} else {
		ord.price = ord.midPrice * (1 + bro.offset)
	}
	ord.volume = getVolume(ord.price, bro.lowLimit, bro.highLimit, bro.base, bro.quote)
	ord.status = "requested"
}

func saveOrder(db *sql.DB, o *order) {
	sqlStmt := `
		INSERT INTO 'order'
			(brokerId, status, orderId, volume, completed, price, tstamp)
			VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING userRef`
	result, err := db.Exec(sqlStmt, o.bro.id, o.status, o.orderId, o.volume,
		o.completed, o.price, time.Now().UnixMilli())
	if err != nil {
		log.Fatal("Couldn't insert order: ", err)
	}
	o.userRef, err = result.LastInsertId()
	if err != nil {
		log.Fatal("Couldn't get order ID: ", err)
	}
}

func updateOrder(db *sql.DB, openOrd kws.OpenOrder, oid string) {
	sqlStmt := `UPDATE 'order' SET orderId=$1 WHERE userRef=$2 AND orderId=''`
	_, err := db.Exec(sqlStmt, oid, openOrd.UserRef)
	if err != nil {
		log.Fatal("Couldn't update id for order: ", openOrd.UserRef)
	}
	sqlStmt = `UPDATE 'order' SET status=$1 WHERE userRef=$2`
	_, err = db.Exec(sqlStmt, openOrd.Status, openOrd.UserRef)
	if err != nil {
		log.Fatal("Couldn't update status for order: ", openOrd.UserRef)
	}
}

func getOrderById(db *sql.DB, acc *account, oid string) *order {
	o := &order{}
	var ts int64
	var broId int64
	sqlStmt := `SELECT * FROM 'order' WHERE orderId=$1`
	if err := db.QueryRow(sqlStmt, oid).Scan(&o.userRef, &broId,
		&o.status, &o.orderId, &o.volume, &o.completed, &o.price,
		&ts); err != nil {
		if err != sql.ErrNoRows {
			log.Debug(err.Error())
		}
		return nil
	} else {
		o.tstamp = time.UnixMilli(ts)
		var ok bool
		o.bro, ok = acc.brokers[broId]
		if !ok {
			log.Info("Kraken order doesn't match any active broker belonging to account.")
			return nil
		}
	}
	return o
}

func getOpenOrderIds(db *sql.DB, bid int64) (oids []string) {
	oids = []string{}
	sqlStmt := `SELECT orderId FROM 'order' WHERE status='open' AND brokerId=$1`
	rows, err := db.Query(sqlStmt, bid)
	if err != nil {
		log.Fatalf("Couldn't get open orders for broker id: %v", bid)
	}
	defer rows.Close()
	for rows.Next() {
		var oid string
		if err := rows.Scan(&oid); err != nil {
			log.Fatal("Couldn't scan open order.")
		}
		oids = append(oids, oid)
	}
	return
}
