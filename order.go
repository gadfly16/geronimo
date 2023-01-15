package main

import (
	"database/sql"
	"time"

	log "github.com/sirupsen/logrus"
)

type order struct {
	userRef  int64
	brokerId int64
	status   string
	orderId  string
	amount   float64
	price    float64
	midPrice float64
	tstamp   time.Time
}

func saveOrder(db *sql.DB, o *order) {
	sqlStmt := `
		INSERT INTO 'order'
			(brokerId, status, orderId, amount, price, tstamp)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING userRef`
	result, err := db.Exec(sqlStmt, o.brokerId, o.status, o.orderId, o.amount,
		o.price, time.Now().UnixMilli())
	if err != nil {
		log.Fatal("Couldn't insert order: ", err)
	}
	o.userRef, err = result.LastInsertId()
	if err != nil {
		log.Fatal("Couldn't get order ID: ", err)
	}
}

func getLastOrder(db *sql.DB, b broker) *order {
	o := &order{}
	var ts int64
	sqlStmt := `SELECT * FROM 'order'
		WHERE brokerId = $1
		ORDER BY tstamp DESC LIMIT 1`
	if err := db.QueryRow(sqlStmt, b.id).Scan(&o.userRef, &o.brokerId, &o.status,
		&o.orderId, &o.amount, &o.price, &ts); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Fatal("Couldn't get last order: ", err)
	}
	o.tstamp = time.UnixMilli(ts)
	return o
}
