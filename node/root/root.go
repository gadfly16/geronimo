package root

import (
	"fmt"
	"log/slog"

	"github.com/gadfly16/geronimo/msg"
	"github.com/gadfly16/geronimo/node"
	"gorm.io/gorm"
)

type RootParms struct {
	ID       int
	NodeID   int
	WorkDir  string
	DbPath   string
	HTTPAddr string
	DbKey    string
}

type RootNode struct {
	node.Head
	Parms RootParms
}

var msgHandlers = map[msg.Kind]func(*RootNode, *msg.Msg) error{
	msg.UpdateKind: updateHandler,
}

func (n *RootNode) run() {
	for m := range n.In {
		if err := msgHandlers[m.Kind](n, m); err != nil {
			m.Resp <- &msg.Msg{Kind: msg.ErrorKind}
		}

		slog.Info("Message received.", "node", n.Name, "MsgKind", m.Kind)
		m.Resp <- &msg.Msg{Kind: msg.OKKind}
	}
}

func updateHandler(n *RootNode, m *msg.Msg) (err error) {
	if v, ok := m.Payload["parms"]; ok {
		if n.Parms, ok = v.(RootParms); !ok {
			return fmt.Errorf("update failed: wrong parms type %T", n.Parms)
		}
	}
	node.Db.Transaction(func(tx *gorm.DB) error {
		return nil
	})
	return
}

func (n *RootNode) Create() (err error) {
	err = node.Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&n.Head).Error; err != nil {
			return err
		}
		n.Parms.NodeID = n.Head.ID
		if err := tx.Create(&n.Parms).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return
	}
	go n.run()
	return
}
