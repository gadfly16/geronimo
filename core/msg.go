package core

import (
	"encoding/json"
	"io"

	"github.com/coder/websocket"
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
	TreeNodeCreateMsgKind
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
	TreeNodeCreateMsgKind:  "TreeNodeCreate",
	GetChildMsgKind:        "GetChild",
	InitGUIMsgKind:         "InitGUI",
	DeleteChildMsgKind:     "DeleteChild",
	TreeNodeDeleteMsgKind:  "TreeNodeDelete",
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

type E struct{}

type DC chan E

type Pipe chan *Msg

type Tag struct {
	ID       int
	Kind     Kind
	Name     string
	Node     Pipe
	ParentID int
	OwnerID  int
}

type Msg struct {
	Kind    MsgKind
	UserID  int
	Admin   bool
	Payload any

	resp Pipe
}

type renameChildPL struct {
	Name    string
	NewName string
}

type CreatePL struct {
	Kind    Kind
	Name    string
	Payload any
}

type NewTreeNodePL struct {
	ID       int
	Name     string
	Kind     Kind
	ParentID int
	OwnerID  int
}

type InitGUIPL struct {
	Conn  *websocket.Conn
	Done  DC
	Admin bool
}

func (p Pipe) MarshalJSON() ([]byte, error) {
	return json.Marshal("Pipe")
}

func (q *Msg) Answer(m *Msg) {
	q.resp <- m
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
	m.resp = make(Pipe)
	t <- &m
	return *<-m.resp
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
		m.Payload = &CreatePL{}
	case RenameChildMsgKind:
		m.Payload = &renameChildPL{}
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
