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

	err = tree.RunGUIClient(uid, c)
	if err != nil {
		slog.Error("GUI couldn't start client.", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
