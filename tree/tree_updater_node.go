package tree

import (
	"fmt"
	"log/slog"
)

type TreeUpdaterNode struct {
	*Head
	treeSubGuis map[*Tag]map[*Tag]E
}

func (t *TreeUpdaterNode) loadBody(h *Head) (n Node, err error) {
	return nil, fmt.Errorf("loading TreeUpdater node is not permitted")
}

func (n *TreeUpdaterNode) create(_ []any) (*Tag, error) {
	n.Head.ID = -NextID()
	n.Head.initNew()
	n.treeSubGuis = make(map[*Tag]map[*Tag]E)
	go n.run()
	return n.Head.Tag, nil
}

func (n *TreeUpdaterNode) run() {
	defer close(n.Head.In)
	defer Tree.RemoveNode(n.Head.ID)

	slog.Debug("TU node starting up.", "node", n.Head.path)
	for q := range n.Head.In {
		a := handleMsg(n, q)
		if a != nil && a.Kind == M_Stop {
			// Sink unsubscribe messages
			slog.Debug("TU number os subs on sinking.", "ngsubs", len(n.Head.guiSubs), "ntsguis", len(n.treeSubGuis))
			nsg := 0
			for _, o := range n.treeSubGuis {
				nsg += len(o)
			}
			for range nsg + len(n.Head.guiSubs) {
				q := <-n.Head.In
				slog.Debug("SINK of TreeUpdater reveived msg.", "node", n.Head.path, "kind", q.KindName())
			}
			q.AnswerMsg(a)
			break
		}
	}

	slog.Info("Stopped TreeUpdater node.", "node", n.path)
}

func (n *TreeUpdaterNode) subscribe(pl []any) {
	if pl[0].(SK) != SK_Tree_Updates {
		slog.Error("TU received an unhandled subscribe.", "sk", pl[0].(SK))
		return
	}
	st := pl[1].(*Tag)
	ot := pl[2].(*Tag)
	if ot.Admin {
		ot = SystemUser
	}
	_, ok := n.treeSubGuis[ot]
	if !ok {
		n.treeSubGuis[ot] = make(map[*Tag]E, 0)
	}
	n.treeSubGuis[ot][st] = E{}
	slog.Debug("TU registered new GUI.", "ntsguis", len(n.treeSubGuis), "owner", st.ID)
}

func (n *TreeUpdaterNode) unsubscribe(pl []any) {
	if pl[0].(SK) != SK_Tree_Updates {
		slog.Error("TU received an unhandled subscribe.", "sk", pl[0].(SK))
		return
	}
	st := pl[1].(*Tag)
	ot := pl[2].(*Tag)
	if ot.Admin {
		ot = SystemUser
	}
	delete(n.treeSubGuis[ot], st)
	if len(n.treeSubGuis[ot]) == 0 {
		delete(n.treeSubGuis, ot)
	}
	slog.Debug("TU unregistered GUI from updates.", "ntsguis", len(n.treeSubGuis))
}

func (tu *TreeUpdaterNode) refreshTree(q *Msg) {
	pl := q.Payload.([]any)
	ot := pl[3].(*Tag)
	for g := range tu.treeSubGuis[SystemUser] {
		g.NotifyMsg(q)
	}
	for g := range tu.treeSubGuis[ot] {
		g.NotifyMsg(q)
	}
	slog.Debug("TU forwarded refresh to affected GUIs.")
}
