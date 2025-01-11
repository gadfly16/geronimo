package core

import (
	"fmt"
	"log/slog"
)

func init() {
	nodeMsgHandlers[TreeUpdaterKind] = map[MsgKind]func(Node, *Msg) *Msg{
		GetDisplayMsgKind:      treeUpdaterGetDisplayHandler,
		SubscribeTreeMsgKind:   subscribeTreeHandler,
		UnsubscribeTreeMsgKind: unsubscribeTreeHandler,
		TreeNodeRenameMsgKind:  treeNodeRenameHandler,
		TreeNodeCreateMsgKind:  treeNodeCreateHandler,
		TreeNodeDeleteMsgKind:  treeNodeDeleteHandler,
		// msg.UpdateKind:   rootUpdateHandler,
		// msg.GetParmsKind: rootGetParmsHandler,
	}
}

type TreeUpdaterNode struct {
	*Head
	treeSubGuis map[int]map[Pipe]bool
}

func (t *TreeUpdaterNode) loadBody(h *Head) (n Node, err error) {
	return nil, fmt.Errorf("loading TreeUpdater node is not permitted")
}

func (n *TreeUpdaterNode) run() {
	defer close(n.Head.In)
	defer Tree.RemoveNode(n.Head.ID)

	slog.Debug("Running TreeUpdater node.", "node", n.Head.path)
	for q := range n.Head.In {
		a := n.Head.handleMsg(n, q)
		if a != nil && a.Kind == StoppedMsgKind {
			// Sink unsubscribe messages
			slog.Debug("TU number os subs on sinking.", "ngsubs", len(n.Head.guiSubs), "ntsguis", len(n.treeSubGuis))
			nsg := 0
			for _, o := range n.treeSubGuis {
				nsg += len(o)
			}
			for range nsg + len(n.Head.guiSubs) {
				q := <-n.Head.In
				slog.Debug("SINK of TreeUpdater reveived msg.", "node", n.Head.path, "kind", q.KindName())
				q.Answer(&OKMsg)
				slog.Debug("SINK of TreeUpdater answered msg.", "node", n.Head.path, "kind", q.KindName())
			}
			q.Answer(a)
			break
		}
	}

	slog.Info("Stopped TreeUpdater node.", "node", n.path)
}

func (n *TreeUpdaterNode) create(_ any) (in Pipe, err error) {
	n.Head.ID = -NextID()
	n.Head.initNew()
	n.treeSubGuis = make(map[int]map[Pipe]bool)
	go n.run()
	return n.Head.In, nil
}

func treeUpdaterGetDisplayHandler(ni Node, _ *Msg) *Msg {
	n := ni.(*TreeUpdaterNode)
	d := n.Head.display()
	r := &Msg{
		Kind:    DisplayMsgKind,
		Payload: d,
	}
	return r
}

func subscribeTreeHandler(ni Node, m *Msg) *Msg {
	n := ni.(*TreeUpdaterNode)
	t := m.Payload.(Tag)
	_, ok := n.treeSubGuis[t.ID]
	if !ok {
		n.treeSubGuis[t.ID] = make(map[Pipe]bool, 0)
	}
	n.treeSubGuis[t.ID][t.Node] = true
	slog.Debug("TU registered new GUI for updates.",
		"ntsguis", len(n.treeSubGuis), "owner", t.ID, "in", t.Node)
	return &OKMsg
}

func unsubscribeTreeHandler(ni Node, m *Msg) *Msg {
	n := ni.(*TreeUpdaterNode)
	t := m.Payload.(Tag)
	delete(n.treeSubGuis[t.ID], t.Node)
	slog.Debug("TU unregistered GUI from updates.", "ntsguis", len(n.treeSubGuis))
	return &OKMsg
}

func treeNodeRenameHandler(ni Node, m *Msg) *Msg {
	n := ni.(*TreeUpdaterNode)
	h := m.Payload.(*Tag)
	for g := range n.treeSubGuis[h.OwnerID] {
		g.Notify(*m)
		slog.Debug("TU notified user GUIs about node rename.", "GUI", h.OwnerID)
	}
	for g := range n.treeSubGuis[0] {
		g.Notify(*m)
		slog.Debug("TU notified admin GUIs about node rename.", "GUI", h.OwnerID)
	}
	slog.Debug("TU handled tree node rename.", "user", h.OwnerID)
	return nil
}

func treeNodeCreateHandler(ni Node, m *Msg) *Msg {
	n := ni.(*TreeUpdaterNode)
	nnpl := m.Payload.(*NewTreeNodePL)
	for g := range n.treeSubGuis[nnpl.OwnerID] {
		g.Notify(*m)
		slog.Debug("TU notified user GUIs about new node.", "GUI", nnpl.OwnerID)
	}
	for g := range n.treeSubGuis[0] {
		g.Notify(*m)
		slog.Debug("TU notified admin GUIs about new node.", "GUI", nnpl.OwnerID)
	}
	slog.Debug("TU handled node creation.", "user", nnpl.OwnerID)
	return nil
}

func treeNodeDeleteHandler(ni Node, m *Msg) *Msg {
	n := ni.(*TreeUpdaterNode)
	nt := m.Payload.(*Tag)
	for g := range n.treeSubGuis[nt.OwnerID] {
		g.Notify(*m)
		slog.Debug("TU notified user GUIs about node deletion.", "GUI", nt.OwnerID)
	}
	for g := range n.treeSubGuis[0] {
		g.Notify(*m)
		slog.Debug("TU notified admin GUI about node deletion.", "GUI", nt.OwnerID)
	}
	slog.Debug("TU handled node deletion.", "user", nt.OwnerID)
	return nil
}
