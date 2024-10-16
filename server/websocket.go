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
	WSMsg_Credentials = "Credentials"
	WSMsg_Subscribe   = "Subscribe"
	WSMsq_Unsubscribe = "Unsubscribe"
	WSMsg_Update      = "Update"
)

type guiClient struct {
	id     int64
	otp    string
	conn   *websocket.Conn
	in     msg.Pipe
	userID int
}

type wsmsg struct {
	Kind   string
	OTP    string
	GUIID  int64
	NodeID int
}

func newGuiClient(conn *websocket.Conn, uid int) (client *guiClient) {
	client = &guiClient{
		conn:   conn,
		id:     node.NextID(),
		otp:    generateOTP(),
		in:     make(msg.Pipe),
		userID: uid,
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

func (gui *guiClient) sendMessage(msg *wsmsg) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = wsjson.Write(ctx, gui.conn, msg)
	if err != nil {
		return err
	}
	return nil
}

func (gui *guiClient) receiveMessage() (msg *wsmsg, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = wsjson.Read(ctx, gui.conn, &msg)
	return
}

func (gui *guiClient) run() {
eventLoop:
	for {
		wm, err := gui.receiveMessage()
		if err != nil {
			switch e := err.(type) {
			case *websocket.CloseError:
				if e.Code == 1000 {
					slog.Info("Client closed connection gracefully.")
				} else {
					slog.Error("Client closed abruptly: ", "error", err)
				}
				break eventLoop
			default:
				slog.Error("Error during receiving ws message:", "error", err)
				continue eventLoop
			}
		}
		if wm.GUIID != gui.id || wm.OTP != gui.otp {
			slog.Error("wrong GUI client credentials")
			continue eventLoop
		}
		switch wm.Kind {
		case WSMsg_Subscribe:
			node, ok := node.Tree.Nodes[wm.NodeID]
			if !ok {
				slog.Error("subscribing to nonexisting node", "node_id", wm.NodeID)
				continue eventLoop
			}
			m := msg.Msg{
				Kind:    msg.SubscribeKind,
				Payload: gui.id,
				UserID:  gui.userID,
			}
			node.Ask(m)
		case WSMsq_Unsubscribe:
			node, ok := node.Tree.Nodes[wm.NodeID]
			if !ok {
				slog.Error("unsubscribing from nonexisting node", "node_id", wm.NodeID)
				continue eventLoop
			}
			m := msg.Msg{
				Kind:    msg.UnsubscribeKind,
				Payload: gui.id,
				UserID:  gui.userID,
			}
			node.Ask(m)
		default:
			slog.Error("unknown websocket message", "wsmsg_kind", wm.Kind)
			continue eventLoop
		}
	}
	slog.Debug("Stopped reading messages for client: ", "gui", gui.id)
}
