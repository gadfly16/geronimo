package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/gadfly16/geronimo/core"
)

const (
	WSMsg_Credentials int = iota
	WSMsg_Subscribe
	WSMsq_Unsubscribe
	WSMsg_Update
	WSMsg_Error
	WSMsg_ClientShutdown
	WSMsg_Heartbeat
	WSMsg_TreeNodeRename
)

type GUIClient struct {
	id     int
	otp    string
	conn   *websocket.Conn
	in     core.Pipe
	wsin   chan wsmsg
	userID int
	admin  bool
	subs   map[int]bool
}

type wsmsg struct {
	Kind     int
	OTP      string
	GUIID    int
	NodeID   int
	NodeName string
}

func newGuiClient(conn *websocket.Conn, uid int, admin bool) (client *GUIClient) {
	client = &GUIClient{
		conn:   conn,
		id:     core.NextID(),
		otp:    generateOTP(),
		in:     make(core.Pipe),
		wsin:   make(chan wsmsg),
		userID: uid,
		admin:  admin,
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

	gui := newGuiClient(c, uid, cls.Admin)
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
	otp, _ := core.GenerateSecret(16)
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

func (gui *GUIClient) receiver() {
	var msg wsmsg
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		err := wsjson.Read(ctx, gui.conn, &msg)
		if err != nil {
			slog.Error("websocket read error", "error", err)
			gui.wsin <- wsmsg{Kind: WSMsg_Error}
			break
		}
		if msg.GUIID != gui.id || msg.OTP != gui.otp {
			slog.Error("wrong websocket credentials")
			gui.wsin <- wsmsg{Kind: WSMsg_Error}
			break
		}
		gui.wsin <- msg
	}
	slog.Debug("GUI ws receiver stopped", "gui", gui.id)
}

func (gui *GUIClient) run() {
	euid := gui.userID
	if gui.admin {
		euid = 0
	}
	core.Tree.TreeUpdater.Ask(core.Msg{
		Kind: core.SubscribeMsgKind,
		Payload: core.SubscribePayload{
			ID:   euid,
			Node: gui.in,
		},
	})

	go gui.receiver()
	slog.Debug("GUI started", "gui", gui.id)
out:
	for {
		select {
		case wm := <-gui.wsin:
			switch wm.Kind {
			case WSMsg_Error:
				break out
			case WSMsg_Subscribe:
				n, ok := core.Tree.GetNode(wm.NodeID)
				if !ok {
					slog.Error("subscribing to nonexisting node", "node_id", wm.NodeID)
					break out
				}
				n.Ask(core.Msg{
					Kind: core.SubscribeMsgKind,
					Payload: core.SubscribePayload{
						ID:   gui.id,
						Node: gui.in,
					},
					UserID: gui.userID,
				})
				gui.subs[wm.NodeID] = true
				slog.Debug("subscribed to node", "node_id", wm.NodeID)
			case WSMsq_Unsubscribe:
				n, ok := core.Tree.GetNode(wm.NodeID)
				if !ok {
					slog.Error("unsubscribing from nonexisting node", "node_id", wm.NodeID)
					break out
				}
				n.Ask(core.Msg{
					Kind:    core.UnsubscribeMsgKind,
					Payload: gui.id,
					UserID:  gui.userID,
				})
				delete(gui.subs, wm.NodeID)
				slog.Debug("unsubscribed to node", "node_id", wm.NodeID)
			case WSMsg_Heartbeat:
				err := gui.sendMessage(&wsmsg{Kind: WSMsg_Heartbeat})
				if err != nil {
					slog.Error("Couldn't send client credentials, closing connection", "error", err)
					break out
				}
			default:
				slog.Error("GUI unknown websocket message", "wsmsg_kind", wm.Kind)
				break out
			}
		case m := <-gui.in:
			switch m.Kind {
			case core.NodeUpdateMsgKind:
				nid := m.Payload.(int)
				err := gui.sendMessage(&wsmsg{
					Kind:   WSMsg_Update,
					NodeID: nid,
				})
				if err != nil {
					slog.Error("GUI couldn't send update msg", "error", err)
					break out
				}
				slog.Debug("GUI sent update to client", "gui", gui.id, "node_id", nid)
			case core.TreeNodeRenameMsgKind:
				slog.Debug("GUI received a tree node rename msg", "msg", m)
				h := m.Payload.(core.Head)
				err := gui.sendMessage(&wsmsg{
					Kind:     WSMsg_TreeNodeRename,
					NodeID:   h.ID,
					NodeName: h.Name,
				})
				if err != nil {
					slog.Error("GUI couldn't send tree update msg", "error", err)
					break out
				}
				slog.Debug("GUI sent tree update to client", "gui", gui.id)
			}
		}
	}
	slog.Debug("GUI stopped reading messages for client: ", "gui", gui.id)
	for nid := range gui.subs {
		n, ok := core.Tree.GetNode(nid)
		if !ok {
			slog.Error("GUI unsubscribing from nonexisting node", "node_id", nid)
		}
		n.Ask(core.Msg{
			Kind:    core.UnsubscribeMsgKind,
			Payload: gui.id,
			UserID:  gui.userID,
		})
	}
	slog.Debug("GUI unsubscribed from nodes: ", "gui", gui.id)

	core.Tree.TreeUpdater.Ask(core.Msg{
		Kind: core.UnsubscribeMsgKind,
		Payload: core.SubscribePayload{
			ID:   gui.userID,
			Node: gui.in,
		},
	})
}
