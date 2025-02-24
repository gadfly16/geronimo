package tree

import (
	"log/slog"

	"gorm.io/gorm"
)

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
		a := handleMsg(n, q)
		if a != nil && a.Kind == M_Stop {
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

func (n *GroupNode) allowedChildren(nk NK) bool {
	return nk != NK_Root && nk != NK_User
}
