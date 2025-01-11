package core

import (
	"log/slog"

	"gorm.io/gorm"
)

func init() {
	nodeMsgHandlers[GroupKind] = map[MsgKind]func(Node, *Msg) *Msg{
		GetDisplayMsgKind: groupGetDisplayHandler,
		// msg.UpdateKind:   rootUpdateHandler,
		// msg.GetParmsKind: rootGetParmsHandler,
	}
}

type GroupNode struct {
	*Head
}

func (t *GroupNode) loadBody(h *Head) (n Node, err error) {
	// h.In = make(Pipe)
	gn := &GroupNode{
		Head: h,
	}

	return gn, nil
}

func (n *GroupNode) run() {
	defer close(n.Head.In)
	defer Tree.RemoveNode(n.Head.ID)

	slog.Debug("Running Group node.", "node", n.Head.path)
	for q := range n.Head.In {
		a := n.Head.handleMsg(n, q)
		if a != nil && a.Kind == StoppedMsgKind {
			// Sink unsubscribe messages
			for range len(n.Head.guiSubs) {
				q := <-n.Head.In
				slog.Debug("SINK of Group reveived msg.", "node", n.Head.path, "kind", q.KindName())
				q.Answer(&OKMsg)
				slog.Debug("SINK of Group answered msg.", "node", n.Head.path, "kind", q.KindName())
			}
			q.Answer(a)
			break
		}
	}

	slog.Info("Stopped Group node.", "node", n.path)
}

func (n *GroupNode) create(_ any) (in Pipe, err error) {
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
	go n.run()
	return n.Head.In, nil
}

func groupGetDisplayHandler(ni Node, _ *Msg) *Msg {
	n := ni.(*GroupNode)
	d := n.Head.display()
	// d["Parms"] = display{
	// 	"Display Name": n.Parms.DisplayName,
	// }
	r := &Msg{
		Kind:    DisplayMsgKind,
		Payload: d,
	}
	return r
}
