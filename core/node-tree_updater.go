package core

import (
	"log/slog"
)

func init() {
	nodeMsgHandlers[TreeUpdaterKind] = map[MsgKind]func(Node, *Msg) *Msg{
		// AuthUserMsgKind:   groupAuthUserHandler,
		GetDisplayMsgKind:     treeUpdaterGetDisplayHandler,
		SubscribeMsgKind:      treeUpdaterSubscribeHandler,
		UnsubscribeMsgKind:    treeUpdaterUnsubscribeHandler,
		TreeNodeRenameMsgKind: treeUpdaterTreeNodeRenameHandler,
		// msg.UpdateKind:   rootUpdateHandler,
		// msg.GetParmsKind: rootGetParmsHandler,
	}
}

type TreeUpdaterNode struct {
	*Head
	guis map[int]map[Pipe]bool
}

func (t *TreeUpdaterNode) loadBody(h *Head) (n Node, err error) {
	// h.In = make(Pipe)
	gn := &TreeUpdaterNode{
		Head: h,
	}

	return gn, nil
}

func (n *TreeUpdaterNode) run() {
	slog.Info("Running TreeUpdater node.", "name", n.Head.Name)
	for q := range n.Head.In {
		a := n.Head.handleMsg(n, q)
		if a != nil && a.Kind == StoppedMsgKind {
			break
		}
	}
	slog.Info("Stopped TreeUpdater node.", "node", n.path)
}

func (n *TreeUpdaterNode) create(p *Head) (in Pipe, err error) {
	n.Head.ID = -NextID()
	n.Head.OwnerID = p.OwnerID
	n.Head.path = p.path + "/" + n.Head.Name
	n.Head.initNew()
	n.guis = make(map[int]map[Pipe]bool)
	slog.Info("Created TreeUpdater node.", "node", n.Head.path)
	go n.run()
	return n.Head.In, nil
}

func treeUpdaterGetDisplayHandler(ni Node, _ *Msg) *Msg {
	n := ni.(*TreeUpdaterNode)
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

func treeUpdaterSubscribeHandler(ni Node, m *Msg) *Msg {
	n := ni.(*TreeUpdaterNode)
	t := m.Payload.(Tag)
	_, ok := n.guis[t.ID]
	if !ok {
		n.guis[t.ID] = make(map[Pipe]bool, 0)
	}
	n.guis[t.ID][t.Node] = true
	slog.Debug("TU registered new GUI for updates.", "guis", n.guis)
	return &OKMsg
}

func treeUpdaterUnsubscribeHandler(ni Node, m *Msg) *Msg {
	n := ni.(*TreeUpdaterNode)
	t := m.Payload.(Tag)
	delete(n.guis[t.ID], t.Node)
	slog.Debug("TU unregistered GUI from updates.", "guis", n.guis)
	return &OKMsg
}

func treeUpdaterTreeNodeRenameHandler(ni Node, m *Msg) *Msg {
	n := ni.(*TreeUpdaterNode)
	h := m.Payload.(Head)
	for g := range n.guis[h.OwnerID] {
		g.Notify(*m)
		slog.Debug("TU notified user GUIs about tree node rename.", "GUI", h.OwnerID)
	}
	for g := range n.guis[0] {
		g.Notify(*m)
		slog.Debug("TU notified admin GUIs about tree node rename.", "GUI", h.OwnerID)
	}
	slog.Debug("TU handled tree node rename.", "user", h.OwnerID)
	return nil
}
