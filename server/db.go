package server

import (
	"errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

func createDB(s Settings) error {
	if FileExists(s.DBPath) {
		return errors.New("database already exists")
	}

	log.Debug("Creating settings database.")

	db, err := gorm.Open(sqlite.Open(s.DBPath), &gorm.Config{})
	if err != nil {
		return err
	}

	db.AutoMigrate(
		&Node{},
		&Checkpoint{},
		&Broker{},
		&Account{},
		&User{},
	)

	log.Infoln("Created database: ", s.DBPath)
	return nil
}
