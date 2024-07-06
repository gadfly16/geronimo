package node

import (
	"errors"
	"log/slog"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Db *gorm.DB

func ConnectDB(path string) (err error) {
	Db, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	return
}

func CloseDB() (err error) {
	slog.Info("Closing state db connection.")
	sqldb, err := Db.DB()
	if err != nil {
		return
	}
	return sqldb.Close()
}

func InitDb(path string) error {
	if FileExists(path) {
		return errors.New("database already exists")
	}

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return err
	}

	db.AutoMigrate(
		&Head{},
		RootParms{},
		UserParms{},
	)

	slog.Info("State database created.", "path", path)
	return nil
}
