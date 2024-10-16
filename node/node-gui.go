package node

import (
	"log/slog"

	"github.com/gadfly16/geronimo/msg"
	"gorm.io/gorm"
)

func init() {
	nodeMsgHandlers[GUIKind] = map[msg.Kind]func(Node, *msg.Msg) *msg.Msg{
		// msg.UpdateKind:   GUIUpdateHandler,
		// msg.GetParmsKind: GUIGetParmsHandler,
	}
}

type GUINode struct {
	*Head
}

func (t *GUINode) loadBody(h *Head) (n Node, err error) {
	gn := &GroupNode{
		Head: h,
	}
	return gn, nil
}

func (n *GUINode) run() {
	slog.Info("Running GUI node.", "name", n.Head.Name)
	for q := range n.Head.In {
		slog.Info("Message received.", "node", n.path, "kind", q.KindName())
		r := n.Head.handleMsg(n, q)
		q.Answer(r)
		slog.Info("Message answered.", "node", n.path, "kind", r.KindName())
		if r.Kind == msg.StoppedKind {
			break
		}
	}
	slog.Info("Stopped Group node.", "node", n.path)
}

func (n *GUINode) create(p *Head) (in msg.Pipe, err error) {
	n.OwnerID = p.OwnerID
	n.Head.path = p.path + "/" + n.Head.Name
	err = Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&n.Head).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return
	}
	n.Head.initNew()
	slog.Info("Created Group node.", "node", n.Head.path)
	go n.run()
	return n.Head.In, nil
}

func (n *GUINode) getDisplay() (d display) {
	d = display{}
	return
}
