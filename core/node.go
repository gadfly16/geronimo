package core

import (
	"time"
)

type Node interface {
	create(any) (Pipe, error)
	loadBody(*Head) (Node, error)
	run()

	// These methods are common to all Nodes, they are defined on Head which is
	// embedded to every Node's struct, therefore declarations are in this file.
	getID() int
	getName() string
	setName(string)
	getPath() string
	setPath(string)
	setParentID(int)
	setKind(Kind)
	setOwnerID(int)
	kindName() string
}

type ParmModel struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	HeadID    int
}

type H map[string]interface{}

func (h *Head) getID() int {
	return h.ID
}

func (h *Head) getName() string {
	return h.Name
}

func (h *Head) setName(n string) {
	h.Name = n
}

func (h *Head) getPath() string {
	return h.path
}

func (h *Head) setPath(p string) {
	h.path = p
}

func (h *Head) setParentID(pid int) {
	h.ParentID = pid
}

func (h *Head) setKind(k Kind) {
	h.Kind = k
}

func (h *Head) setOwnerID(oid int) {
	h.OwnerID = oid
}

func (h *Head) kindName() string {
	return h.KindName()
}
