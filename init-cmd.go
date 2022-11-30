package main

import (
	"database/sql"
	"flag"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func init() {
	commands["init"] = initCommand
}

func initCommand() {
	log.Debug("Running 'init' command.")

	initFlags := flag.NewFlagSet("init", flag.ExitOnError)
	initFlags.Parse(flag.Args()[1:])

	if dbExists() {
		log.Fatal("Settings database already exists.")
	}

	log.Debug("Creating settings database.")

	db, err := sql.Open("sqlite3", "./settings.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
		PRAGMA foreign_keys = ON ;

		CREATE TABLE account (
			id INTEGER PRIMARY KEY,
			name TEXT UNIQUE,
			apiPublicKey TEXT,
			apiPrivateKey TEXT
		);

		CREATE TABLE brokerHead (
			id INTEGER PRIMARY KEY,
			name TEXT UNIQUE,
			accountId INTEGER,
			FOREIGN KEY (accountId) REFERENCES account (id)
		);
			
		CREATE TABLE brokerSetting (
			brokerId INTEGER,
			status INTEGER,
			minWait REAL,
			maxWait REAL,
			highLimit REAL,
			lowLimit REAL,
			delta REAL,
			offset REAL,
			modt DATETIME DEFAULT(STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
			FOREIGN KEY (brokerId) REFERENCES broker (id)
		);

		CREATE TABLE brokerBalance (
			brokerId INTEGER,
			base REAL,
			quote REAL,
			modt DATETIME DEFAULT(STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
			FOREIGN KEY (brokerId) REFERENCES broker (id)
		);
		
		CREATE VIEW broker AS 
			SELECT
				bh.id, bh.name, bh.accountId, bs.status,
				bs.minWait, bs.maxWait, bs.highLimit, bs.lowLimit, bs.delta, bs.offset,
				bb.base, bb.quote FROM brokerHead bh
			JOIN brokerSetting bs ON bh.id=bs.brokerId
			JOIN brokerBalance bb ON bh.id=bb.brokerId
			WHERE bs.modt = (SELECT max(modt) FROM brokerSetting bs2 WHERE bs2.brokerId = bh.id)
				AND bb.modt = (SELECT max(modt) FROM brokerBalance bb2 WHERE bb2.brokerId = bh.id)
		; `

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal("%q: %s\n", err, sqlStmt)
	}
}
