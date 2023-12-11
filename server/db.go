package server

import (
	"errors"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

func dbExists(s Settings) bool {
	if _, err := os.Stat(s.SettingsDbPath); err == nil {
		return true
	} else if !errors.Is(err, os.ErrNotExist) {
		log.Fatal("Couldn't stat settings database.")
	}
	return false
}

func CreateDB(s Settings) error {
	if dbExists(s) {
		return errors.New("database already exists")
	}

	log.Debug("Creating settings database.")

	db, err := gorm.Open(sqlite.Open(s.SettingsDbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	db.AutoMigrate(&Checkpoint{}, &Broker{}, &Account{})

	log.Infoln("Created database: ", s.SettingsDbPath)
	return nil
}
