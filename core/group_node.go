package core

import (
	"log/slog"

	mk "github.com/gadfly16/geronimo/msgKinds"
	"gorm.io/gorm"
)

func init() {
	nodeMsgHandlers[GroupKind] = map[mk.MK]func(Node, *Msg) *Msg{
		mk.GetDisplay: groupGetDisplayHandler,
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

	slog.Debug("GROUP node starting up.", "node", n.Head.path)
	for q := range n.Head.In {
		a := n.Head.handleMsg(n, q)
		if a != nil && a.Kind == mk.Stopped {
			// Sink unsubscribe messages
			for range len(n.Head.guiSubs) {
				q := <-n.Head.In
				slog.Debug("SINK of Group reveived msg.", "node", n.Head.path, "kind", q.KindName())
				q.AnswerOK()
				slog.Debug("SINK of Group answered msg.", "node", n.Head.path, "kind", q.KindName())
			}
			q.AnswerMsg(a)
			break
		}
	}

	slog.Info("Stopped Group node.", "node", n.path)
}

func (n *GroupNode) create(_ []any) (_ *Tag, err error) {
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
	return n.Head.Tag, nil
}

func groupGetDisplayHandler(ni Node, _ *Msg) *Msg {
	n := ni.(*GroupNode)
	d := n.Head.display()
	// d["Parms"] = display{
	// 	"Display Name": n.Parms.DisplayName,
	// }
	r := &Msg{
		Kind:    mk.Display,
		Payload: d,
	}
	return r
}
