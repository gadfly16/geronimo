package node

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Db *gorm.DB

func ConnectDB(path string) (err error) {
	Db, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	return
}
