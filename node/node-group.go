package node

import (
	"log/slog"

	"github.com/gadfly16/geronimo/msg"
	"gorm.io/gorm"
)

func init() {
	nodeMsgHandlers[GroupKind] = map[msg.Kind]func(Node, *msg.Msg) *msg.Msg{
		// msg.UpdateKind:   rootUpdateHandler,
		// msg.GetParmsKind: rootGetParmsHandler,
	}
}

type GroupNode struct {
	*Head
}

func (t *GroupNode) loadBody(h *Head) (n Node, err error) {
	h.In = make(msg.Pipe)
	gn := &GroupNode{
		Head: h,
	}

	return gn, nil
}

func (n *GroupNode) run() {
	slog.Info("Running Group node.", "name", n.Head.Name)
	for m := range n.Head.In {
		slog.Info("Message received.", "node", n.Head.Name, "kind", m.KindName())
		r := n.Head.handleMsg(n, m)
		r.Answer(m)
		slog.Info("Message answered.", "node", n.Head.Name, "kind", r.KindName())
		if r.Kind == msg.StoppedKind {
			break
		}
	}
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
