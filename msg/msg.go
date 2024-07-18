package msg

const (
	OKKind Kind = iota
	ErrorKind
	StopKind
	StoppedKind
	UpdateKind
	ParmsKind
	GetParmsKind
	CreateKind
	AuthUserKind
	GetTreeKind
	TreeKind
	GetCopyKind
	GetDetailKind
	DetailKind
)

var kindNames = map[Kind]string{
	OKKind:        "OK",
	ErrorKind:     "Error",
	StopKind:      "Stop",
	StoppedKind:   "Stopped",
	UpdateKind:    "Update",
	ParmsKind:     "Parms",
	GetParmsKind:  "GetParms",
	CreateKind:    "Create",
	AuthUserKind:  "AuthUser",
	GetTreeKind:   "GetTree",
	TreeKind:      "Tree",
	GetCopyKind:   "GetCopy",
	GetDetailKind: "GetDetail",
	DetailKind:    "Detail",
}

var (
	OK        = Msg{Kind: OKKind}
	Stop      = Msg{Kind: StopKind}
	GetParms  = Msg{Kind: GetParmsKind}
	GetCopy   = Msg{Kind: GetCopyKind}
	GetTree   = Msg{Kind: GetTreeKind}
	GetDetail = Msg{Kind: GetDetailKind}
)

type Pipe chan *Msg

type Kind = int

type Msg struct {
	Kind    Kind
	Payload any
	Resp    Pipe

	auth struct {
		uid   int
		admin bool
	}
}

func (m *Msg) Answer(q *Msg) {
	q.Resp <- m
}

func NewErrorMsg(err error) *Msg {
	return &Msg{
		Kind:    ErrorKind,
		Payload: err.Error(),
	}
}

func (m *Msg) ErrorMsg() string {
	return m.Payload.(string)
}

func (m *Msg) KindName() string {
	return kindNames[m.Kind]
}

func (t Pipe) Ask(m *Msg) *Msg {
	m.Resp = make(Pipe)
	t <- m
	return <-m.Resp
}
