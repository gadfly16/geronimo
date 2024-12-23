package core

import (
	"time"
)

type Node interface {
	create(*Head) (Pipe, error)
	loadBody(*Head) (Node, error)
	run()

	// These methods are common to all Nodes, they are defined on Head which is
	// embedded to every Node's struct, therefore declarations are in this file.
	getName() string
	setName(string)
	setParentID(int)
	getPath() string
}

type ParmModel struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	HeadID    int
}

type display map[string]interface{}

func (h *Head) getName() string {
	return h.Name
}

func (h *Head) setName(n string) {
	h.Name = n
}

func (h *Head) setParentID(pid int) {
	h.ParentID = pid
}

func (h *Head) getPath() string {
	return h.path
}
