package core

import (
	"encoding/json"
	"io"
)

type MsgKind = int

const (
	OKMsgKind MsgKind = iota
	ErrorMsgKind
	StopMsgKind
	StoppedMsgKind
	UpdateMsgKind
	ParmsMsgKind
	GetParmsMsgKind
	CreateChildMsgKind
	AuthUserMsgKind
	GetTreeMsgKind
	TreeMsgKind
	GetCopyMsgKind
	GetDisplayMsgKind
	DisplayMsgKind
	SubscribeMsgKind
	UnsubscribeMsgKind
	NodeUpdateMsgKind
	RenameMsgKind
	TreeNodeRenameMsgKind // PL: [t *Tag, nm string, ot *Tag]
	UpdatePathMsgKind
	RenameChildMsgKind
	CreateUserMsgKind
	SubscribeTreeMsgKind
	UnsubscribeTreeMsgKind
	TreeNodeCreateMsgKind // PL: [t *Tag, nm string, ot *Tag]
	GetChildMsgKind
	InitGUIMsgKind
	DeleteChildMsgKind
	TreeNodeDeleteMsgKind
)

var MsgKindNames = map[MsgKind]string{
	OKMsgKind:              "OK",
	ErrorMsgKind:           "Error",
	StopMsgKind:            "Stop",
	StoppedMsgKind:         "Stopped",
	UpdateMsgKind:          "Update",
	ParmsMsgKind:           "Parms",
	GetParmsMsgKind:        "GetParms",
	CreateChildMsgKind:     "CreateChild",
	AuthUserMsgKind:        "AuthUser",
	GetTreeMsgKind:         "GetTree",
	TreeMsgKind:            "Tree",
	GetCopyMsgKind:         "GetCopy",
	GetDisplayMsgKind:      "GetDisplay",
	DisplayMsgKind:         "Display",
	SubscribeMsgKind:       "Subscribe",
	UnsubscribeMsgKind:     "Unsubscribe",
	NodeUpdateMsgKind:      "NodeUpdate",
	RenameMsgKind:          "Rename",
	TreeNodeRenameMsgKind:  "TreeNodeRename",
	UpdatePathMsgKind:      "UpdatePath",
	RenameChildMsgKind:     "RenameChild",
	CreateUserMsgKind:      "CreateUser",
	SubscribeTreeMsgKind:   "SubscribeTree",
	UnsubscribeTreeMsgKind: "UnsubscribeTree",
	TreeNodeCreateMsgKind:  "TreeNodeCreate",
	GetChildMsgKind:        "GetChild",
	InitGUIMsgKind:         "InitGUI",
	DeleteChildMsgKind:     "DeleteChild",
	TreeNodeDeleteMsgKind:  "TreeNodeDelete",
}

var OKMsg = Msg{Kind: OKMsgKind}

type E struct{}

type DC chan E

type Pipe chan *Msg

func (p Pipe) MarshalJSON() ([]byte, error) {
	return json.Marshal("Pipe")
}

type Msg struct {
	Kind    MsgKind
	User    *Tag
	Payload any

	resp Pipe
}

// type CreateChildPL struct {
// 	Kind    Kind
// 	Name    string
// 	Payload any
// }

// type InitGUIPL struct {
// 	Conn  *websocket.Conn
// 	Done  DC
// 	Admin bool
// }

func (t *Tag) Ask(mk MsgKind, u *Tag, pl ...any) Msg {
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
	return *<-m.resp
}

func (t *Tag) AskMsg(m *Msg) Msg {
	m.resp = make(Pipe)
	t.In <- m
	return *<-m.resp
}

func (t *Tag) Notify(mk MsgKind, u *Tag, pl ...any) {
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

func (q *Msg) Answer(mk MsgKind, pl ...any) {
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
	q.resp <- &OKMsg
}

func NewErrorMsg(err error) *Msg {
	return &Msg{
		Kind:    ErrorMsgKind,
		Payload: err.Error(),
	}
}

func (m *Msg) ErrorMsg() string {
	return m.Payload.(string)
}

func (m *Msg) KindName() string {
	return MsgKindNames[m.Kind]
}

func UnmarshalMsg(mk MsgKind, b io.ReadCloser) (m *Msg, err error) {
	m = &Msg{}
	switch mk {
	case UpdateMsgKind:
		m.Payload = H{}
	case CreateChildMsgKind, RenameChildMsgKind:
		m.Payload = []any{}
	default:
		m.Payload = nil
	}
	if m.Payload != nil {
		d := json.NewDecoder(b)
		err = d.Decode(&m.Payload)
	}
	if err != nil {
		return nil, err
	}
	m.Kind = mk
	return
}
