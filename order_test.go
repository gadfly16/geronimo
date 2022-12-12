package main

import (
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	databaseFlag = "./testing.db"
	os.Remove(databaseFlag)
	createDB()
	m.Run()
}

func TestOrderTime(t *testing.T) {
	db := openDB()
	defer db.Close()

	bro := broker{}

	ord := &order{}
	ord.brokerId = bro.id
	ts := time.Now()
	ord.tstamp = ts
	saveOrder(db, ord)

	ord = lastOrder(db, bro)
	ts = time.UnixMilli(ts.UnixMilli())
	if ord.tstamp != ts {
		t.Errorf("Order's timestamp doesn't match: got %s, expected %s .",
			ord.tstamp, ts)
	}
}

func TestNoLastOrder(t *testing.T) {
	db := openDB()
	defer db.Close()

	sqlStmt := `DELETE FROM 'order'`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Fatal("Couldn't delete orders from test db.")
	}

	bro := broker{}

	ord := lastOrder(db, bro)
	if ord != nil {
		t.Errorf("Order returned when no orders are in db.")
	}
}
