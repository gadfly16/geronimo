package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/gadfly16/geronimo/msg"
	"github.com/gadfly16/geronimo/node"
)

const (
	WSMsg_Credentials int = iota
	WSMsg_Subscribe
	WSMsq_Unsubscribe
	WSMsg_Update
	WSMsg_Error
	WSMsg_ClientShutdown
	WSMsg_Heartbeat
)

type GUIClient struct {
	id     int64
	otp    string
	conn   *websocket.Conn
	in     msg.Pipe
	userID int
	subs   map[int]bool
}

type wsmsg struct {
	Kind   int
	OTP    string
	GUIID  int64
	NodeID int
}

func newGuiClient(conn *websocket.Conn, uid int) (client *GUIClient) {
	client = &GUIClient{
		conn:   conn,
		id:     node.NextID(),
		otp:    generateOTP(),
		in:     make(msg.Pipe),
		userID: uid,
		subs:   make(map[int]bool),
	}
	return
}

func socketHandler(w http.ResponseWriter, q *http.Request) {
	cls := q.Context().Value(ctxClaims).(*claims)
	uid, err := strconv.Atoi(cls.Subject)
	if err != nil {
		slog.Error("invalid user ID")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c, err := websocket.Accept(w, q, nil)
	if err != nil {
		slog.Error("Can't establish websocket connection")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer c.CloseNow()

	gui := newGuiClient(c, uid)
	msg := &wsmsg{
		Kind:  WSMsg_Credentials,
		GUIID: gui.id,
		OTP:   gui.otp,
	}
	err = gui.sendMessage(msg)
	if err != nil {
		slog.Error("Couldn't send client credentials, closing connection", "error", err)
		return
	}

	_, err = gui.receiveMessage()
	if err != nil {
		slog.Error("Error during client id affirmation", "error", err)
	}
	slog.Debug("GUI affirmation received", "gui_id", gui.id)

	gui.run()
}

func generateOTP() string {
	otp, _ := node.GenerateSecret(16)
	for i, b := range otp {
		otp[i] = b%94 + 33
	}
	return string(otp)
}

func (gui *GUIClient) sendMessage(msg *wsmsg) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = wsjson.Write(ctx, gui.conn, msg)
	if err != nil {
		return err
	}
	return nil
}

func (gui *GUIClient) receiveMessage() (msg *wsmsg, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = wsjson.Read(ctx, gui.conn, &msg)
	return
}

func (gui *GUIClient) receiver(mc chan wsmsg) {
	var msg wsmsg
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		err := wsjson.Read(ctx, gui.conn, &msg)
		if err != nil {
			slog.Error("websocket read error", "error", err)
			mc <- wsmsg{Kind: WSMsg_Error}
			break
		}
		if msg.GUIID != gui.id || msg.OTP != gui.otp {
			slog.Error("wrong websocket credentials")
			mc <- wsmsg{Kind: WSMsg_Error}
			break
		}
		mc <- msg
	}
}

func (gui *GUIClient) run() {
	rmsgc := make(chan wsmsg)
	go gui.receiver(rmsgc)

out:
	for {
		select {
		case wm := <-rmsgc:
			switch wm.Kind {
			case WSMsg_Error:
				break out
			case WSMsg_Subscribe:
				n, ok := node.Tree.Nodes[wm.NodeID]
				if !ok {
					slog.Error("subscribing to nonexisting node", "node_id", wm.NodeID)
					break out
				}
				m := msg.Msg{
					Kind: msg.SubscribeKind,
					Payload: node.SubscribePayload{
						GUIID: gui.id,
						GUIIn: gui.in,
					},
					UserID: gui.userID,
				}
				n.Ask(m)
				gui.subs[wm.NodeID] = true
				slog.Debug("subscribed to node", "node_id", wm.NodeID)
			case WSMsq_Unsubscribe:
				n, ok := node.Tree.Nodes[wm.NodeID]
				if !ok {
					slog.Error("unsubscribing from nonexisting node", "node_id", wm.NodeID)
					break out
				}
				m := msg.Msg{
					Kind:    msg.UnsubscribeKind,
					Payload: gui.id,
					UserID:  gui.userID,
				}
				n.Ask(m)
				delete(gui.subs, wm.NodeID)
				slog.Debug("unsubscribed to node", "node_id", wm.NodeID)
			case WSMsg_Heartbeat:
				err := gui.sendMessage(&wsmsg{Kind: WSMsg_Heartbeat})
				if err != nil {
					slog.Error("Couldn't send client credentials, closing connection", "error", err)
					break out
				}
				// slog.Debug("Heartbeat sent.", "gui_id", gui.id)
			default:
				slog.Error("GUI unknown websocket message", "wsmsg_kind", wm.Kind)
				break out
			}
		case m := <-gui.in:
			slog.Debug("GUI received msg from node", "msg", m)
			nid := m.Payload.(int)
			wsm := &wsmsg{
				Kind:   WSMsg_Update,
				NodeID: nid,
			}
			err := gui.sendMessage(wsm)
			if err != nil {
				slog.Error("GUI couldn't send update msg", "error", err)
				break out
			}
			slog.Debug("GUI sent update to client", "gui", gui.id, "node_id", nid)
		}
	}
	slog.Debug("GUI stopped reading messages for client: ", "gui", gui.id)
	for nid := range gui.subs {
		n, ok := node.Tree.Nodes[nid]
		if !ok {
			slog.Error("GUI unsubscribing from nonexisting node", "node_id", nid)
		}
		n.Ask(msg.Msg{
			Kind:    msg.UnsubscribeKind,
			Payload: gui.id,
			UserID:  gui.userID,
		})
	}
	slog.Debug("GUI unsubscribed from nodes: ", "gui", gui.id)
}
