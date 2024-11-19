package core

import (
	"time"
)

type Node interface {
	create(*Head) (Pipe, error)
	loadBody(*Head) (Node, error)
	run()
	getName() string
	setName(string)
	setParentID(int)
}

type ParmModel struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	HeadID    int
}

type display map[string]interface{}
