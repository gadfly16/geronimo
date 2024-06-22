package node

import (
	"time"

	"gorm.io/gorm"

	"github.com/gadfly16/geronimo/msg"
)

const (
	RootKind = iota
	GroupKind
	NodeAccount
)

type Param struct{}

// type Credit struct {
// 	currency int
// 	amount   decimal.Decimal
// }

type Kind = int

type Head struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name     string
	Kind     Kind
	ParentID int
	In       msg.Pipe `gorm:"-"`

	parent   msg.Pipe
	children map[string]msg.Pipe

	// cache struct {
	// 	valid bool
	// }
}

type Node interface {
	run() msg.Pipe
	path() error
	create(Kind)
}

type Group struct {
	Head
}

type Account struct {
	Head
	Exchange string
}

func (n *Head) load() error {
	return nil
}

func (n *Head) save() error {
	return nil
}

func (n *Head) path() string {
	return n.Name
}

func (n *Head) idByPath() int {
	nn := &Account{}
	nn.path()
	return 1
}

// func (n *NodeHead) send(m *Msg) {
// 	n.Nerd.In <- m
// }

// func (n *NodeHead) wait(m *Msg) (r *Msg) {
// 	rc := make(chan *Msg)
// 	n.Nerd.In <- m
// 	return <-rc
// }

// func (n *NodeHead) run() {
// 	for m := range n.Nerd.In {
// 		slog.Debug("messge received", n.Nerd.Name, m)
// 	}
// }
