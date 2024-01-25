package server

import (
	"errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

func createDB(s Settings) error {
	if fileExists(s.DBName) {
		return errors.New("database already exists")
	}

	log.Debug("Creating settings database.")

	db, err := gorm.Open(sqlite.Open(s.DBName), &gorm.Config{})
	if err != nil {
		return err
	}

	db.AutoMigrate(&Checkpoint{}, &Broker{}, &Account{}, &User{})

	log.Infoln("Created database: ", s.DBName)
	return nil
}
