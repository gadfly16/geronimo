package node

import (
	"time"

	"github.com/gadfly16/geronimo/msg"
)

type Node interface {
	create(*Head) (msg.Pipe, error)
	loadBody(*Head) (Node, error)
	run()
	name() string
	setParentID(int)
}

type Group struct {
	Head
}

type Account struct {
	Head
	Exchange string
}

type ParmModel struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	HeadID    int
}

type display map[string]interface{}
