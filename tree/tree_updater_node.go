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

func (n *TreeUpdaterNode) run() {
	defer close(n.Head.In)
	defer Tree.RemoveNode(n.Head.ID)

	slog.Debug("TU node starting up.", "node", n.Head.path)
	for q := range n.Head.In {
		a := n.Head.handleMsg(n, q)
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
				q.AnswerOK()
				slog.Debug("SINK of TreeUpdater answered msg.", "node", n.Head.path, "kind", q.KindName())
			}
			q.AnswerMsg(a)
			break
		}
	}

	slog.Info("Stopped TreeUpdater node.", "node", n.path)
}

func (n *TreeUpdaterNode) create(_ []any) (*Tag, error) {
	n.Head.ID = -NextID()
	n.Head.initNew()
	n.treeSubGuis = make(map[*Tag]map[*Tag]E)
	go n.run()
	return n.Head.Tag, nil
}

func treeUpdaterGetDisplayHandler(ni Node, _ *Msg) *Msg {
	n := ni.(*TreeUpdaterNode)
	d := n.Head.display()
	r := &Msg{
		Kind:    M_OK,
		Payload: d,
	}
	return r
}

func subscribeTreeHandler(ni Node, q *Msg) *Msg {
	n := ni.(*TreeUpdaterNode)
	pl := q.Payload.([]any)
	st := pl[0].(*Tag)
	ot := pl[1].(*Tag)
	if ot.Admin {
		ot = SystemUser
	}
	_, ok := n.treeSubGuis[ot]
	if !ok {
		n.treeSubGuis[ot] = make(map[*Tag]E, 0)
	}
	n.treeSubGuis[ot][st] = E{}
	slog.Debug("TU registered new GUI for updates.",
		"ntsguis", len(n.treeSubGuis), "owner", st.ID, "in", st.In)
	return oka
}

func unsubscribeTreeHandler(ni Node, q *Msg) *Msg {
	n := ni.(*TreeUpdaterNode)
	pl := q.Payload.([]any)
	st := pl[0].(*Tag)
	ot := pl[1].(*Tag)
	if ot.Admin {
		ot = SystemUser
	}
	delete(n.treeSubGuis[ot], st)
	if len(n.treeSubGuis[ot]) == 0 {
		delete(n.treeSubGuis, ot)
	}
	slog.Debug("TU unregistered GUI from updates.", "ntsguis", len(n.treeSubGuis))
	return oka
}

func treeNodeRenameHandler(ni Node, q *Msg) *Msg {
	n := ni.(*TreeUpdaterNode)
	pl := q.Payload.([]any)
	ot := pl[2].(*Tag) // GUI owner tag
	for g := range n.treeSubGuis[SystemUser] {
		g.NotifyMsg(q)
		slog.Debug("TU notified admin GUIs about node rename.", "GUI", q.User.ID)
	}
	for g := range n.treeSubGuis[ot] {
		g.NotifyMsg(q)
		slog.Debug("TU notified user GUIs about node rename.", "GUI", q.User.ID)
	}
	slog.Debug("TU handled tree node rename.", "user", q.User.ID)
	return nil
}

func treeNodeCreateHandler(ni Node, q *Msg) *Msg {
	n := ni.(*TreeUpdaterNode)
	pl := q.Payload.([]any)
	ot := pl[2].(*Tag)
	for g := range n.treeSubGuis[SystemUser] {
		g.NotifyMsg(q)
		slog.Debug("TU notified admin GUIs about node rename.", "GUI", q.User.ID)
	}
	for g := range n.treeSubGuis[ot] {
		g.NotifyMsg(q)
		slog.Debug("TU notified user GUIs about new node.", "GUI", q.User.ID)
	}
	slog.Debug("TU handled node creation.", "user", q.User.ID)
	return nil
}

func treeNodeDeleteHandler(ni Node, q *Msg) *Msg {
	n := ni.(*TreeUpdaterNode)
	pl := q.Payload.([]any)
	ot := pl[2].(*Tag)
	for g := range n.treeSubGuis[ot] {
		g.NotifyMsg(q)
		slog.Debug("TU notified user GUIs about node deletion.", "GUI", q.User.ID)
	}
	for g := range n.treeSubGuis[SystemUser] {
		g.NotifyMsg(q)
		slog.Debug("TU notified admin GUI about node deletion.", "GUI", q.User.ID)
	}
	slog.Debug("TU handled node deletion.", "user", q.User.ID)
	return nil
}
