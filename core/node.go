package core

import (
	"time"
)

type Node interface {
	create(any) (*Tag, error)
	loadBody(*Head) (Node, error)
	run()

	// These methods are common to all Nodes, they are defined on Head which is
	// embedded to every Node's struct, therefore declarations are in this file.
	getID() NodeID
	getName() string
	setName(string)
	getPath() string
	setPath(string)
	setParentID(*Tag)
	setKind(Kind)
	setOwnerID(*Tag)
	kindName() string
}

type ParmModel struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	HeadID    NodeID
}

type H map[string]interface{}

func (h *Head) getID() NodeID {
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

func (h *Head) setParentID(pt *Tag) {
	h.Parent = pt
	h.ParentID = pt.ID
}

func (h *Head) setKind(k Kind) {
	h.Kind = k
}

func (h *Head) setOwnerID(ot *Tag) {
	h.Owner = ot
}

func (h *Head) kindName() string {
	return h.KindName()
}
