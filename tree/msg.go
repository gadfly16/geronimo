package tree

import (
	"encoding/json"
)

type MK = int

var oka = &Msg{Kind: M_OK}

const (
	M_OK MK = iota
	M_Error

	M_Create
	M_Rename
	M_Delete

	M_Get_Parms
	M_Get_Auth
	M_Get_Tree
	M_Get_Display
	M_Get_Child

	M_Update_Parms

	M_Refresh_Node
	M_Refresh_Tree

	M_Subscribe
	M_Unsubscribe

	M_Stop

	M_UpdatePath
)

// MNames is exported for logging purposes.
var MKNames = map[MK]string{
	M_OK:    "OK",
	M_Error: "Error",

	M_Create: "Create",
	M_Rename: "Rename",
	M_Delete: "Delete",

	M_Get_Parms:   "Get_Parms",
	M_Get_Auth:    "Get_Auth",
	M_Get_Tree:    "Get_Tree",
	M_Get_Display: "Get_Display",
	M_Get_Child:   "Get_Child",

	M_Update_Parms: "Update_Parms",

	M_Refresh_Node: "Refresh_Node",
	M_Refresh_Tree: "Refresh_Tree",

	M_Subscribe:   "Subscribe",
	M_Unsubscribe: "Unsubscribe",

	M_Stop: "Stop",

	M_UpdatePath: "UpdatePath",
}

func (p Pipe) MarshalJSON() ([]byte, error) {
	return json.Marshal("Pipe")
}

type Msg struct {
	Kind    MK
	User    *Tag
	Payload any

	resp Pipe
}

func (t *Tag) Ask(u *Tag, mk MK, pl ...any) *Msg {
	// For sake of comfort, if there's only one payload, we'll use it directly.
	var epl any = pl
	if len(pl) == 1 {
		epl = pl[0]
	}
	m := &Msg{
		Kind:    mk,
		Payload: epl,
		User:    u,
		resp:    make(Pipe),
	}
	t.In <- m
	return <-m.resp
}

func (t *Tag) AskMsg(m *Msg) Msg {
	m.resp = make(Pipe)
	t.In <- m
	return *<-m.resp
}

func (t *Tag) Notify(u *Tag, mk MK, pl ...any) {
	var epl any = pl
	if len(pl) == 1 {
		epl = pl[0]
	}
	m := &Msg{
		Kind:    mk,
		Payload: epl,
		User:    u,
	}
	t.In <- m
}

func (t *Tag) NotifyMsg(m *Msg) {
	t.In <- m
}

func (q *Msg) Answer(mk MK, pl ...any) {
	var epl any = pl
	if len(pl) == 1 {
		epl = pl[0]
	}
	a := &Msg{
		Kind:    mk,
		Payload: epl,
		User:    q.User,
	}
	q.resp <- a
}

func (q *Msg) AnswerMsg(a *Msg) {
	q.resp <- a
}

func (q *Msg) AnswerOK() {
	q.resp <- oka
}

func NewErrorMsg(err error) *Msg {
	return &Msg{
		Kind:    M_Error,
		Payload: err,
	}
}

func (m *Msg) KindName() string {
	return MKNames[m.Kind]
}
