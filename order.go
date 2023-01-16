package main

import (
	"database/sql"
	"time"

	kws "github.com/aopoltorzhicky/go_kraken/websocket"
	log "github.com/sirupsen/logrus"
)

type order struct {
	userRef   int64
	brokerId  int64
	status    string
	orderId   string
	volume    float64
	completed float64
	price     float64
	midPrice  float64
	tstamp    time.Time
}

func saveOrder(db *sql.DB, o *order) {
	o.status = "requested"
	sqlStmt := `
		INSERT INTO 'order'
			(brokerId, status, orderId, volume, completed, price, tstamp)
			VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING userRef`
	result, err := db.Exec(sqlStmt, o.brokerId, o.status, o.orderId, o.volume,
		o.completed, o.price, time.Now().UnixMilli())
	if err != nil {
		log.Fatal("Couldn't insert order: ", err)
	}
	o.userRef, err = result.LastInsertId()
	if err != nil {
		log.Fatal("Couldn't get order ID: ", err)
	}
}

func getLastOrder(db *sql.DB, bro *broker) *order {
	o := &order{}
	var ts int64
	sqlStmt := `SELECT * FROM 'order'
		WHERE brokerId = $1 AND completed > 0
		ORDER BY tstamp DESC LIMIT 1`
	if err := db.QueryRow(sqlStmt, bro.id).Scan(&o.userRef, &o.brokerId, &o.status,
		&o.orderId, &o.volume, &o.completed, &o.price, &ts); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Fatal("Couldn't get last order: ", err)
	}
	o.tstamp = time.UnixMilli(ts)
	return o
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

func getOrderById(db *sql.DB, oid string) (o *order) {
	o = &order{}
	var ts int64
	sqlStmt := `SELECT * FROM 'order' WHERE orderId=$1`
	if err := db.QueryRow(sqlStmt, oid).Scan(&o.userRef, &o.brokerId,
		&o.status, &o.orderId, &o.volume, &o.completed, &o.price,
		&ts); err != nil {
		if err != sql.ErrNoRows {
			log.Debug(err.Error())
		}
		o = nil
	} else {
		o.tstamp = time.UnixMilli(ts)
	}
	return
}

func getOpenOrderIds(db *sql.DB, bid int64) (oids []string) {
	oids = []string{}
	sqlStmt := `SELECT orderId FROM 'order' WHERE status='open' AND brokerId=$1`
	rows, err := db.Query(sqlStmt, bid)
	if err != nil {
		log.Fatalf("Couldn't get open orders for broker: ", bid)
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