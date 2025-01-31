package tree

import (
	"time"
)

type Node interface {
	create(pl []any) (*Tag, error)
	loadBody(*Head) (Node, error)
	run()

	// These methods are common to all Nodes, they are defined on Head which is
	// embedded to every Node's struct, therefore declarations are in this file.
	head() *Head
	kindName() string
}

type ParmModel struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	HeadID    NodeID
}

type H map[string]interface{}

func (h *Head) head() *Head {
	return h
}

func (h *Head) kindName() string {
	return h.KindName()
}
