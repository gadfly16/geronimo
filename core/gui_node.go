package core

import (
	"fmt"
	"log/slog"
)

func init() {
	nodeMsgHandlers[GUIKind] = map[MsgKind]func(Node, *Msg) *Msg{
		GetDisplayMsgKind: groupGetDisplayHandler,
		// msg.UpdateKind:   rootUpdateHandler,
		// msg.GetParmsKind: rootGetParmsHandler,
	}
}

type GUINode struct {
	*Head
	IP string
}

func (t *GUINode) loadBody(h *Head) (n Node, err error) {
	return nil, fmt.Errorf("loading GUI node is not permitted")
}

func (n *GUINode) run() {
	slog.Info("Running GUI node.", "name", n.Head.Name)
	for q := range n.Head.In {
		a := n.Head.handleMsg(n, q)
		if a != nil && a.Kind == StoppedMsgKind {
			break
		}
	}
	slog.Info("Stopped GUI node.", "node", n.path)
}

func (n *GUINode) create() (in Pipe, err error) {
	n.Head.ID = -NextID()
	n.Head.Name = fmt.Sprintf("GUI_%d", n.Head.ID)
	n.Head.initNew()
	slog.Info("Created GUI node.", "node", n.Head.path)
	go n.run()
	return n.Head.In, nil
}

func guiGetDisplayHandler(ni Node, _ *Msg) *Msg {
	n := ni.(*GUINode)
	d := n.Head.display()
	r := &Msg{
		Kind:    DisplayMsgKind,
		Payload: d,
	}
	return r
}
