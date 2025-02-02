package tree

import (
	"fmt"
	"log/slog"
)

// Message handler function
type HF func(Node, *Msg) *Msg

var msgHandlers []HF = []HF{
	M_Create: createHandler,
	M_Stop:   stopHandler,
}

func handleMsg(n Node, q *Msg) (a *Msg) {
	if q.User != n.head().Owner && !q.User.Admin {
		return NewErrorMsg(fmt.Errorf("unathorized message"))
	}
	slog.Debug("MSG received.", "node", n.head().path, "qk", q.KindName())
	hf := msgHandlers[q.Kind]
	if hf == nil {
		return NewErrorMsg(fmt.Errorf("no handler for msg kind: %v", q.Kind))
	}
	a = hf(n, q)
	if a != nil && a.Kind != M_Stop {
		q.AnswerMsg(a)
	}
	slog.Debug("MSG handled.", "node", n.head().path, "qk", q.KindName(), "ak", a.KindName())
	return
}

func stopHandler(n Node, q *Msg) (a *Msg) {
	n.head().askChildrenMsg(q)
	return &Msg{Kind: M_Stop, Payload: n.head().ID}
}

func createHandler(n Node, q *Msg) (a *Msg) {
	pl := q.Payload.([]any)
	nnk := pl[0].(NK)
	nnm := pl[1].(string)
	if nnk == NK_Root || nnk == NK_User {
		return NewErrorMsg(fmt.Errorf("%s kind can not be created", NKNames[nnk]))
	}
	if nnm == "" {
		nnm = ("New" + NKNames[nnk])
	}
	if _, ok := n.head().children[nnm]; ok {
		return NewErrorMsg(fmt.Errorf("node '%s' already exists", nnm))
	}
	nn := newNode(nnk)
	if nn == nil {
		return NewErrorMsg(fmt.Errorf("node kind '%s' not implemented yet", NKNames[nnk]))
	}
	nn.head().Kind = nnk
	nn.head().Name = nnm
	nn.head().ParentID = n.head().Tag.ID
	nn.head().Parent = n.head().Tag
	nn.head().Owner = n.head().Owner

	nnt, err := nn.create(pl[2:])
	if err != nil {
		return NewErrorMsg(err)
	}
	//Name might have been changed by create.
	nnm = nn.head().Name
	n.head().children[nnm] = nnt
	nn.head().path = n.head().path + "/" + nnm

	if Tree.Sys.TreeUpdater != nil {
		Tree.Sys.TreeUpdater.Notify(SystemUser, M_Update_Tree, M_Create, nnt, nnm, n.head().Owner)
	}

	slog.Debug("NODE created.", "node", nn.head().path, "kind", nn.kindName())
	return &Msg{Kind: M_OK, Payload: nnt}
}

// func stopHandler(n Node, q *Msg) (a *Msg) {
// 	n.head().askChildrenMsg(q)
// 	return &Msg{Kind: M_Stop, Payload: []any{n.head().ID}}
// }

// func updateCreateHandler(h *Head, q *Msg) (a *Msg) {
// 	// pl := q.Payload.([]any)
// 	return oka
// }

// func getParmsHandler(h *Head, q *Msg) (a *Msg) {
// 	return oka
// }

// func getAuthHandler(h *Head, q *Msg) (a *Msg) {
// 	return oka
// }
