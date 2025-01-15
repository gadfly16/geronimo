package core

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func init() {
	nodeMsgHandlers[GUIKind] = map[MsgKind]func(Node, *Msg) *Msg{
		GetDisplayMsgKind:     guiGetDisplayHandler,
		NodeUpdateMsgKind:     guiNodeUpdateHandler,
		TreeNodeRenameMsgKind: guiTreeNodeRenameHandler,
		TreeNodeCreateMsgKind: guiTreeNodeCreateHandler,
		TreeNodeDeleteMsgKind: guiTreeNodeDeleteHandler,
		// msg.UpdateKind:   rootUpdateHandler,
		// msg.GetParmsKind: rootGetParmsHandler,
	}
}

type GUINode struct {
	*Head
	// IP string
	conn     *websocket.Conn
	wsr      chan wsMsg
	otp      string
	admin    bool
	subNodes map[NodeID]E
	httpDone DC
}

func (t *GUINode) loadBody(h *Head) (n Node, err error) {
	return nil, fmt.Errorf("loading GUI node is not permitted")
}

func (n *GUINode) run() {
	defer close(n.Head.In)
	defer close(n.httpDone)
	defer Tree.RemoveNode(n.Head.ID)

	slog.Debug("GUI node starting up.", "node", n.Head.path)
	// Authenticating websocket connection
	err := n.sendWSMessage(&wsMsg{
		Kind:  CredentialsWsMsgKind,
		GUIID: n.Head.ID,
		OTP:   n.otp,
	})
	if err != nil {
		slog.Error("GUI couldn't send credentials.", "error", err)
		return
	}
	cr, err := n.receiveWSMessage()
	if err != nil || cr.OTP != n.otp {
		slog.Error("Error during client id affirmation", "error", err)
	}
	slog.Debug("GUI credential affirmation received.", "gui_id", n.Head.ID)
	// Subscribe for tree updates
	Tree.Sys.TreeUpdater.Ask(SubscribeTreeMsgKind, SystemUser, n.Tag, n.Owner)
	// Start satelites
	guiCtx, stopGuiCtx := context.WithCancel(context.Background())
	defer stopGuiCtx()
	wsrdone := make(DC)
	go n.wsReceiver(guiCtx, wsrdone)

out:
	for {
		select {
		case wm := <-n.wsr:
			switch wm.Kind {
			case SubscribeWsMsgKind:
				if wm.NodeID != n.ID {
					sn, ok := Tree.GetNode(wm.NodeID)
					if !ok {
						slog.Error("subscribing to nonexisting node", "node_id", wm.NodeID)
						break out
					}
					sn.Notify(SubscribeMsgKind, n.Owner, n.Tag)
				}
				n.subNodes[wm.NodeID] = E{}
			case UnsubscribeWsMsgKind:
				if wm.NodeID != n.ID {
					sn, ok := Tree.GetNode(wm.NodeID)
					if !ok {
						slog.Error("unsubscribing from nonexisting node", "node_id", wm.NodeID)
						break out
					}
					sn.Notify(UnsubscribeMsgKind, n.Owner, n.Tag)
				}
				delete(n.subNodes, wm.NodeID)
			case HeartbeatWsMsgKind:
				err := n.sendWSMessage(&wsMsg{Kind: HeartbeatWsMsgKind})
				// slog.Debug("GUI sent heartbeat.", "gui", gui.path)
				if err != nil {
					slog.Error("GUI couldn't send client credentials.", "error", err)
					break out
				}
			default:
				slog.Error("GUI unknown websocket message.", "wsmsg_kind", wm.Kind)
				break out
			}
		case q := <-n.In:
			a := n.Head.handleMsg(n, q)
			if a != nil && a.Kind == StoppedMsgKind {
				// Stop satelites
				// stopGuiCtx()
				n.conn.Close(websocket.StatusServiceRestart, "stopping node")
				<-wsrdone
				// Unsubscribe from nodes
				for nid := range n.subNodes {
					if nid != n.Head.ID {
						go func() {
							sn, ok := Tree.GetNode(nid)
							if !ok {
								slog.Error("GUI unsubscribe from nonexisting node.", "node_id", nid)
								return
							}
							sn.Ask(UnsubscribeMsgKind, n.Owner, n.Tag)
						}()
					}
				}
				// Unsubscribe from tree updater
				Tree.Sys.TreeUpdater.Ask(UnsubscribeTreeMsgKind, SystemUser, n.Tag, n.Owner)
				// Drain unsubscribe messages
				for range len(n.Head.guiSubs) {
					q := <-n.Head.In
					q.AnswerOK()
				}
				q.AnswerMsg(a)
				break out
			}
		}
	}
	slog.Info("Stopped GUI node.", "node", n.Head.path)
}

func (n *GUINode) create(pl any) (*Tag, error) {
	n.Head.ID = -NextID()
	n.Head.Name = fmt.Sprintf("GUI%d", n.Head.ID)

	igpl := pl.(*InitGUIPL)
	n.httpDone = igpl.Done
	n.conn = igpl.Conn
	n.admin = igpl.Admin
	n.wsr = make(chan wsMsg)
	n.subNodes = make(map[NodeID]E)
	n.otp = generateOTP()

	n.Head.initNew()
	go n.run()
	return n.Head.Tag, nil
}

func guiNodeUpdateHandler(guii Node, q *Msg) (a *Msg) {
	gui := guii.(*GUINode)
	t := q.Payload.(*Tag)
	slog.Debug("GUI received a node update msg.", "gui", gui.path, "t", t.ID)
	err := gui.sendWSMessage(&wsMsg{
		Kind:   UpdateWsMsgKind,
		NodeID: t.ID,
	})
	if err != nil {
		slog.Error("GUI couldn't send update msg", "error", err)
		return NewErrorMsg(err)
	}
	slog.Debug("GUI sent node update to client", "gui", gui.path, "tid", t.ID)
	return nil
}

func guiTreeNodeRenameHandler(guii Node, q *Msg) (a *Msg) {
	gui := guii.(*GUINode)
	pl := q.Payload.([]any)
	t := pl[0].(*Tag)
	nm := pl[1].(string)
	slog.Debug("GUI received a tree node rename Msg.", "gui", gui.path, "t", t.ID)
	err := gui.sendWSMessage(&wsMsg{
		Kind:     TreeNodeRenameWsMsgKind,
		NodeID:   t.ID,
		NodeName: nm,
	})
	if err != nil {
		slog.Error("GUI couldn't send tree node rename msg", "err", err)
		return NewErrorMsg(err)
	}
	slog.Debug("GUI sent tree node rename to client", "gui", gui.path, "t", t.ID, "nm", nm)
	return nil
}

func guiTreeNodeDeleteHandler(guii Node, q *Msg) (a *Msg) {
	gui := guii.(*GUINode)
	pl := q.Payload.([]any)
	t := pl[0].(*Tag)
	nm := pl[1].(string)
	slog.Debug("GUI received a tree node delete msg.", "gui", gui.path, "t", t.ID)
	err := gui.sendWSMessage(&wsMsg{
		Kind:         TreeNodeDeleteWsMsgKind,
		NodeParentID: t.Parent.ID,
		NodeName:     nm,
	})
	if err != nil {
		slog.Error("GUI couldn't send node deletion msg.", "error", err)
		return NewErrorMsg(err)
	}
	slog.Debug("GUI sent node deletion to client.", "gui", gui.path, "tid", t.ID, "nm", nm)
	return nil
}

func guiTreeNodeCreateHandler(guii Node, q *Msg) (a *Msg) {
	gui := guii.(*GUINode)
	slog.Debug("GUI received a new tree node msg", "gui", gui.path, "node", gui.path)
	pl := q.Payload.([]any)
	t := pl[0].(*Tag)
	nm := pl[1].(string)
	err := gui.sendWSMessage(&wsMsg{
		Kind:         TreeNodeCreateWsMsgKind,
		NodeID:       t.ID,
		NodeName:     nm,
		NodeKind:     t.Kind,
		NodeParentID: t.Parent.ID,
	})
	if err != nil {
		slog.Error("GUI couldn't send tree create msg", "error", err)
		return NewErrorMsg(err)
	}
	slog.Debug("GUI sent tree node create to client", "gui", gui.path, "tid", t.ID, "nm", nm)
	return nil
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

func (n *GUINode) wsReceiver(ctx context.Context, done DC) {
	var msg wsMsg
	slog.Debug("GUI ws receiver started.", "gui", n.Head.ID)
	for {
		ctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		err := wsjson.Read(ctx, n.conn, &msg)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusServiceRestart {
				slog.Error("GUI ws receiver stopped by node.", "node", n.path, "error", err)
				close(done)
			} else {
				slog.Error("GUI ws read error, exiting.", "node", n.path, "error", err)
				close(done)
				n.Parent.Ask(DeleteChildMsgKind, n.Owner, n.Name)
			}
			break
		}
		if msg.GUIID != n.Head.ID || msg.OTP != n.otp {
			slog.Error("GUI encountered bad ws credentials, exiting.")
			close(done)
			n.Parent.Ask(DeleteChildMsgKind, n.Owner, n.Tag)
			break
		}
		n.wsr <- msg
	}
	slog.Debug("GUI WS Receiver stopped.", "gui", n.Head.ID)
}

func (gui *GUINode) sendWSMessage(q *wsMsg) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return wsjson.Write(ctx, gui.conn, q)
}

func (gui *GUINode) receiveWSMessage() (a *wsMsg, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = wsjson.Read(ctx, gui.conn, &a)
	return
}
