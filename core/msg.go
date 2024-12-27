package core

import (
	"encoding/json"
	"io"
)

const (
	OKMsgKind MsgKind = iota
	ErrorMsgKind
	StopMsgKind
	StoppedMsgKind
	UpdateMsgKind
	ParmsMsgKind
	GetParmsMsgKind
	CreateMsgKind
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
	TreeNodeRenameMsgKind
	UpdatePathMsgKind
	RenameChildMsgKind
	CreateUserMsgKind
	SubscribeTreeMsgKind
	UnsubscribeTreeMsgKind
)

var MsgKindNames = map[MsgKind]string{
	OKMsgKind:              "OK",
	ErrorMsgKind:           "Error",
	StopMsgKind:            "Stop",
	StoppedMsgKind:         "Stopped",
	UpdateMsgKind:          "Update",
	ParmsMsgKind:           "Parms",
	GetParmsMsgKind:        "GetParms",
	CreateMsgKind:          "Create",
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
}

var (
	StopMsg       = Msg{Kind: StopMsgKind}
	OKMsg         = Msg{Kind: OKMsgKind}
	StoppedMsg    = Msg{Kind: StoppedMsgKind}
	GetParmsMsg   = Msg{Kind: GetParmsMsgKind}
	GetCopyMsg    = Msg{Kind: GetCopyMsgKind}
	GetTreeMsg    = Msg{Kind: GetTreeMsgKind}
	GetDisplayMsg = Msg{Kind: GetDisplayMsgKind}
	UpdatedMsg    = Msg{Kind: NodeUpdateMsgKind}
)

type Pipe chan *Msg

type MsgKind = int

type Msg struct {
	Kind    MsgKind
	Payload any
	Resp    Pipe

	UserID int
	Admin  bool
}

type renameChildPayload struct {
	Name    string
	NewName string
}

type CreatePayload struct {
	Kind Kind
	Name string
}

func (p Pipe) MarshalJSON() ([]byte, error) {
	return json.Marshal("Pipe")
}

func (q *Msg) Answer(m *Msg) {
	q.Resp <- m
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

func (t Pipe) Ask(m Msg) Msg {
	m.Resp = make(Pipe)
	t <- &m
	return *<-m.Resp
}

func (t Pipe) Notify(m Msg) {
	t <- &m
}

func UnmarshalMsg(mk MsgKind, b io.ReadCloser) (m *Msg, err error) {
	m = &Msg{}
	switch mk {
	case UpdateMsgKind:
		m.Payload = map[string]interface{}{}
	case CreateMsgKind:
		m.Payload = &CreatePayload{}
	case RenameChildMsgKind:
		m.Payload = &renameChildPayload{}
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
