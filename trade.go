package main

import (
	"database/sql"
	"log"
	"time"
)

type trade struct {
	id      string
	orderId string
	volume  float64
	cost    float64
	fee     float64
	tstamp  time.Time
}

func tradeExists(db *sql.DB, id string) bool {
	sqlStmt := `SELECT count(*) FROM trade WHERE id=$1`
	var count int64
	if err := db.QueryRow(sqlStmt, id).Scan(&count); err != nil {
		log.Fatalf("Couldn't check existence of trade '%v': %v", id, err)
	}
	return count == 1
}
