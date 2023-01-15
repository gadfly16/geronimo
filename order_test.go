package main

import (
	"database/sql"
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

func deleteOrders(db *sql.DB) {
	sqlStmt := `DELETE FROM 'order'`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Fatal("Couldn't delete orders from test db.")
	}
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

	ord = getLastOrder(db, bro)
	ts = time.UnixMilli(ts.UnixMilli())
	if ord.tstamp != ts {
		t.Errorf("Order's timestamp doesn't match: got %s, expected %s .",
			ord.tstamp, ts)
	}
}

func TestOrderId(t *testing.T) {
	db := openDB()
	defer db.Close()

	deleteOrders(db)

	ord := &order{}
	saveOrder(db, ord)

	if ord.userRef == 0 {
		t.Errorf("Order's `userRef` should be other than 0.")
	}
}

func TestNoLastOrder(t *testing.T) {
	db := openDB()
	defer db.Close()

	deleteOrders(db)

	bro := broker{}

	ord := getLastOrder(db, bro)
	if ord != nil {
		t.Errorf("Order returned when no orders are in db.")
	}
}
