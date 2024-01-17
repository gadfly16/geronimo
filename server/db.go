package server

import (
	"errors"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

func dbExists(dbName string) bool {
	if _, err := os.Stat(dbName); err == nil {
		return true
	} else if !errors.Is(err, os.ErrNotExist) {
		log.Fatal("Couldn't stat settings database.")
	}
	return false
}

func createDB(s Settings) error {
	dbName := s.WorkDir + "/state.db"
	if dbExists(dbName) {
		return errors.New("database already exists")
	}

	log.Debug("Creating settings database.")

	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		return err
	}

	db.AutoMigrate(&Checkpoint{}, &Broker{}, &Account{}, &User{})

	log.Infoln("Created database: ", dbName)
	return nil
}
