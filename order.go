package main

import (
	"database/sql"
	"time"

	log "github.com/sirupsen/logrus"
)

type order struct {
	brokerId int64
	status   string
	amount   float64
	price    float64
	midPrice float64
	tstamp   time.Time
}

func saveOrder(db *sql.DB, o *order) {
	sqlStmt := `INSERT INTO 'order' VALUES ($1, $2, $3, $4, $5)`
	_, err := db.Exec(sqlStmt, o.brokerId, o.status, o.amount,
		o.price, o.tstamp.UnixMilli())
	if err != nil {
		log.Fatal("Couldn't insert order: ", err)
	}
}

func getLastOrder(db *sql.DB, b broker) *order {
	o := &order{}
	var ts int64
	sqlStmt := `SELECT * FROM 'order'
		WHERE brokerId = $1
		ORDER BY tstamp DESC LIMIT 1`
	if err := db.QueryRow(sqlStmt, b.id).Scan(&o.brokerId, &o.status,
		&o.amount, &o.price, &ts); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Fatal("Couldn't get last order: ", err)
	}
	o.tstamp = time.UnixMilli(ts)
	return o
}
