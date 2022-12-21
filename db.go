package main

import (
	"database/sql"
	"errors"
	"os"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func dbExists() bool {
	if _, err := os.Stat(databaseFlag); err == nil {
		return true
	} else if !errors.Is(err, os.ErrNotExist) {
		log.Fatal("Couldn't stat settings database.")
	}
	return false
}

func openDB() *sql.DB {
	if !dbExists() {
		log.Fatal("Settings database doesn't exists.")
	}
	db, err := sql.Open("sqlite3", databaseFlag)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func createDB() {
	if dbExists() {
		log.Fatal("Settings database already exists.")
	}

	log.Debug("Creating settings database.")

	db, err := sql.Open("sqlite3", databaseFlag)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
		PRAGMA foreign_keys = ON ;

		CREATE TABLE account (
			id INTEGER PRIMARY KEY,
			name TEXT UNIQUE,
			password TEXT,
			apiPublicKey TEXT,
			apiPrivateKey TEXT
		);

		CREATE TABLE brokerHead (
			id INTEGER PRIMARY KEY,
			accountId INTEGER,
			name TEXT UNIQUE,
			pair TEXT,
			FOREIGN KEY (accountId) REFERENCES account (id)
		);
			
		CREATE TABLE brokerSetting (
			brokerId INTEGER,
			status TEXT,
			minWait REAL,
			maxWait REAL,
			highLimit REAL,
			lowLimit REAL,
			delta REAL,
			offset REAL,
			modt DATETIME DEFAULT(STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
			FOREIGN KEY (brokerId) REFERENCES brokerHead (id)
		);

		CREATE TABLE brokerBalance (
			brokerId INTEGER,
			base REAL,
			quote REAL,
			modt DATETIME DEFAULT(STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
			FOREIGN KEY (brokerId) REFERENCES brokerHead (id)
		);
		
		CREATE VIEW broker AS 
			SELECT
				bh.id, bh.accountId, bh.name, bh.pair, bs.status, bs.minWait,
				bs.maxWait, bs.highLimit, bs.lowLimit, bs.delta, bs.offset,
				bb.base, bb.quote FROM brokerHead bh
			JOIN brokerSetting bs ON bh.id=bs.brokerId
			JOIN brokerBalance bb ON bh.id=bb.brokerId
			WHERE bs.modt = (SELECT max(modt) FROM brokerSetting bs2 WHERE bs2.brokerId = bh.id)
				AND bb.modt = (SELECT max(modt) FROM brokerBalance bb2 WHERE bb2.brokerId = bh.id)
		;
		
		CREATE TABLE 'order' (
			brokerId INTEGER,
			status TEXT,
			amount REAL,
			price REAL,
			tstamp INTEGER,
			FOREIGN KEY (brokerId) REFERENCES brokerHead (id)
		); `

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}
}
