package main

import (
	"database/sql"
	"errors"
	"os"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func dbExists() bool {
	if _, err := os.Stat("./settings.db"); err == nil {
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
	db, err := sql.Open("sqlite3", "./settings.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}
