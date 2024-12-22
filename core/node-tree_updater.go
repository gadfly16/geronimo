package core

import (
	"log/slog"

	"gorm.io/gorm"
)

func init() {
	nodeMsgHandlers[TreeUpdaterKind] = map[MsgKind]func(Node, *Msg) *Msg{
		// AuthUserMsgKind:   groupAuthUserHandler,
		GetDisplayMsgKind: treeUpdaterGetDisplayHandler,
		// msg.UpdateKind:   rootUpdateHandler,
		// msg.GetParmsKind: rootGetParmsHandler,
	}
}

type TreeUpdaterNode struct {
	*Head
}

func (t *TreeUpdaterNode) loadBody(h *Head) (n Node, err error) {
	// h.In = make(Pipe)
	gn := &TreeUpdaterNode{
		Head: h,
	}

	return gn, nil
}

func (n *TreeUpdaterNode) run() {
	slog.Info("Running Group node.", "name", n.Head.Name)
	for q := range n.Head.In {
		slog.Info("Message received.", "node", n.path, "kind", q.KindName())
		a := n.Head.handleMsg(n, q)
		q.Answer(a)
		slog.Info("Message answered.", "node", n.path, "reqKind", q.KindName(), "ansKind", a.KindName())
		if a.Kind == StoppedMsgKind {
			break
		}
	}
	slog.Info("Stopped Group node.", "node", n.path)
}

func (n *TreeUpdaterNode) create(p *Head) (in Pipe, err error) {
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

func treeUpdaterGetDisplayHandler(ni Node, _ *Msg) *Msg {
	n := ni.(*GroupNode)
	d := n.Head.display()
	// d["Parms"] = display{
	// 	"Display Name": n.Parms.DisplayName,
	// 	"Admin":        n.Parms.Admin,
	// }
	// slog.Debug("Display data returned by user node", "displayData", d)
	r := &Msg{
		Kind:    DisplayMsgKind,
		Payload: d,
	}
	return r
}
