package tree

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

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

func (n *GUINode) create(pl []any) (*Tag, error) {
	n.Head.ID = -NextID()
	n.Head.Name = fmt.Sprintf("GUI%d", n.Head.ID)

	conn := pl[0].(*websocket.Conn)
	done := pl[1].(DC)
	admin := pl[2].(bool)
	n.httpDone = done
	n.conn = conn
	n.admin = admin
	n.wsr = make(chan wsMsg)
	n.subNodes = make(map[NodeID]E)
	n.otp = generateOTP()

	n.Head.initNew()
	go n.run()
	return n.Head.Tag, nil
}

func (n *GUINode) run() {
	defer close(n.Head.In)
	defer close(n.httpDone)
	defer Tree.RemoveNode(n.Head.ID)

	slog.Debug("GUI node starting up.", "node", n.Head.path)
	// Authenticating websocket connection
	err := n.sendWSMessage(&wsMsg{
		Kind:  WSM_Credentials,
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
	Tree.Sys.TreeUpdater.Notify(SystemUser, M_Subscribe, SK_Tree_Updates, n.Tag, n.Owner)
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
			case WSM_Subscribe:
				if wm.NodeID != n.ID {
					sn, ok := getNode(wm.NodeID)
					if !ok {
						slog.Error("subscribing to nonexisting node", "node_id", wm.NodeID)
						break out
					}
					sn.Notify(n.Owner, M_Subscribe, SK_Node_Updates, n.Tag)
				}
				n.subNodes[wm.NodeID] = E{}
			case WSM_Unsubscribe:
				if wm.NodeID != n.ID {
					sn, ok := getNode(wm.NodeID)
					if !ok {
						slog.Error("unsubscribing from nonexisting node", "node_id", wm.NodeID)
						break out
					}
					sn.Notify(n.Owner, M_Unsubscribe, SK_Node_Updates, n.Tag)
				}
				delete(n.subNodes, wm.NodeID)
			case WSM_Heartbeat:
				err := n.sendWSMessage(&wsMsg{Kind: WSM_Heartbeat})
				if err != nil {
					slog.Error("GUI couldn't send client credentials.", "error", err)
					break out
				}
			default:
				slog.Error("GUI unknown websocket message.", "wsmsg_kind", wm.Kind)
				break out
			}
		case q := <-n.In:
			a := handleMsg(n, q)
			if a != nil && a.Kind == M_Stop {
				// Stop satelites
				n.conn.Close(websocket.StatusServiceRestart, "stopping node")
				<-wsrdone
				// Unsubscribe from nodes
				for nid := range n.subNodes {
					if nid != n.Head.ID {
						go func() {
							sn, ok := getNode(nid)
							if !ok {
								slog.Error("GUI unsubscribe from nonexisting node.", "node_id", nid)
								return
							}
							sn.Ask(n.Owner, M_Unsubscribe, SK_Node_Updates, n.Tag)
						}()
					}
				}
				// Unsubscribe from tree updater
				Tree.Sys.TreeUpdater.Notify(SystemUser, M_Unsubscribe, SK_Tree_Updates, n.Tag, n.Owner)
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
				n.Parent.Ask(n.Owner, M_Delete, n.Name)
			}
			break
		}
		if msg.GUIID != n.Head.ID || msg.OTP != n.otp {
			slog.Error("GUI encountered bad ws credentials, exiting.")
			close(done)
			n.Parent.Ask(n.Owner, M_Delete, n.Tag)
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

func refreshNodeHandler(n Node, q *Msg) (a *Msg) {
	gui, ok := n.(*GUINode)
	if !ok {
		return NewErrorMsg(fmt.Errorf("%s cant refresh node", n.head().path))
	}
	t := q.Payload.(*Tag)
	err := gui.sendWSMessage(&wsMsg{
		Kind:   WSM_Update,
		NodeID: t.ID,
	})
	if err != nil {
		slog.Error("GUI couldn't send update msg", "error", err)
		return NewErrorMsg(err)
	}
	slog.Debug("GUI sent node update to client", "gui", gui.path, "tid", t.ID)
	return nil
}

func (gui *GUINode) refreshTree(q *Msg) {
	pl := q.Payload.([]any)
	rk := pl[0].(MK)
	t := pl[1].(*Tag)
	nm := pl[2].(string)
	switch rk {
	case M_Create:
		err := gui.sendWSMessage(&wsMsg{
			Kind:         WSM_Created,
			NodeID:       t.ID,
			NodeName:     nm,
			NodeKind:     t.Kind,
			NodeParentID: t.Parent.ID,
		})
		if err != nil {
			slog.Error("GUI couldn't send tree create msg", "error", err)
			return
		}
		slog.Debug("GUI sent tree node create to client", "gui", gui.path, "tid", t.ID, "nm", nm)
	case M_Rename:
		slog.Debug("GUI received a tree node rename Msg.", "gui", gui.path, "t", t.ID)
		err := gui.sendWSMessage(&wsMsg{
			Kind:     WSM_Renamed,
			NodeID:   t.ID,
			NodeName: nm,
		})
		if err != nil {
			slog.Error("GUI couldn't send tree node rename msg", "err", err)
			return
		}
		slog.Debug("GUI sent tree node rename to client", "gui", gui.path, "t", t.ID, "nm", nm)
	case M_Delete:
		slog.Debug("GUI received a tree node delete msg.", "gui", gui.path, "t", t.ID)
		err := gui.sendWSMessage(&wsMsg{
			Kind:         WSM_Deleted,
			NodeParentID: t.Parent.ID,
			NodeName:     nm,
		})
		if err != nil {
			slog.Error("GUI couldn't send node deletion msg.", "error", err)
			return
		}
		slog.Debug("GUI sent node deletion to client.", "gui", gui.path, "tid", t.ID, "nm", nm)
	}
}
