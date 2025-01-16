package server

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/coder/websocket"

	"github.com/gadfly16/geronimo/core"
	mk "github.com/gadfly16/geronimo/msgKinds"
)

func socketHandler(w http.ResponseWriter, q *http.Request) {
	cls := q.Context().Value(ctxClaims).(*claims)
	uid, err := strconv.Atoi(cls.Subject)
	if err != nil {
		slog.Error("Invalid user ID.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c, err := websocket.Accept(w, q, nil)
	if err != nil {
		slog.Error("Can't establish websocket connection.")
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
	u, ok := core.Tree.GetNode(core.NodeID(uid))
	if !ok {
		slog.Error("Can't get User node.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	a := un.Ask(mk.GetChild, u, "GUIs")

	guis := a.Payload.(*core.Tag)
	done := make(core.DC)
	a = guis.Ask(mk.CreateChild, u, core.GUIKind, "", c, done, cls.Admin)

	<-done
}
