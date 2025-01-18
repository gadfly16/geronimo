package server

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/coder/websocket"

	"github.com/gadfly16/geronimo/tree"
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

	un, ok := tree.Tree.GetNode(tree.NodeID(uid))
	if !ok {
		slog.Error("invalid user ID")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	u, ok := tree.Tree.GetNode(tree.NodeID(uid))
	if !ok {
		slog.Error("Can't get User node.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	a := un.Ask(tree.MK_GetChild, u, "GUIs")

	guis := a.Payload.(*tree.Tag)
	done := make(tree.DC)
	a = guis.Ask(tree.MK_CreateChild, u, tree.NK_GUI, "", c, done, cls.Admin)

	<-done
}
