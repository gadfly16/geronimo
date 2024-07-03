package node

import (
	"log/slog"

	"github.com/gadfly16/geronimo/msg"
	"gorm.io/gorm"
)

type GroupNode struct {
	*Head
}

var groupMsgHandlers = map[msg.Kind]func(Node, *msg.Msg) *msg.Msg{
	// msg.UpdateKind:   rootUpdateHandler,
	// msg.GetParmsKind: rootGetParmsHandler,
}

func (t *GroupNode) load(h *Head) (n Node, err error) {
	h.In = make(msg.Pipe)
	gn := &GroupNode{
		Head: h,
	}

	return gn, nil
}

func (n *GroupNode) run() {
	slog.Info("Running Group node.", "name", n.Head.Name)
	for m := range n.Head.In {
		slog.Info("Message received.", "node", n.Head.Name, "MsgKind", m.Kind)
		r := n.Head.commonMsg(m)
		if r == nil {
			r = groupMsgHandlers[m.Kind](n, m)
		}
		r.Answer(m)
		slog.Info("Message answered.", "node", n.Head.Name, "MsgKind", m.Kind)
	}
	n.Head.stopChildren()
	slog.Info("Stopped Group node.", "name", n.Head.Name)
}

func (n *GroupNode) create() (in msg.Pipe, err error) {
	err = Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&n.Head).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return
	}
	n.Head.register()
	slog.Info("Created Group node.", "name", n.Head.Name)
	go n.run()
	return n.Head.In, nil
}
