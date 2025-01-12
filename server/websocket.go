package server

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/coder/websocket"

	"github.com/gadfly16/geronimo/core"
)

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

	un, ok := core.Tree.GetNode(core.NodeID(uid))
	if !ok {
		slog.Error("invalid user ID")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	a := un.Ask(core.Msg{
		Kind:    core.GetChildMsgKind,
		UserID:  core.NodeID(uid),
		Admin:   cls.Admin,
		Payload: "GUIs",
	})

	guis := a.Payload.(core.Pipe)
	done := make(core.DC)
	a = guis.Ask(core.Msg{
		Kind:   core.CreateMsgKind,
		UserID: core.NodeID(uid),
		Admin:  cls.Admin,
		Payload: &core.CreatePL{
			Kind: core.GUIKind,
			Payload: &core.InitGUIPL{
				Conn:  c,
				Done:  done,
				Admin: cls.Admin,
			},
		},
	})

	<-done
}
