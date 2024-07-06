package msg

const (
	OKKind = iota
	ErrorKind
	StopKind
	StoppedKind
	UpdateKind
	ParmsKind
	GetParmsKind
	CreateKind
)

var kindNames = map[Kind]string{
	OKKind:       "OK",
	ErrorKind:    "Error",
	StopKind:     "Stop",
	StoppedKind:  "Stopped",
	UpdateKind:   "Update",
	ParmsKind:    "Parms",
	GetParmsKind: "GetParms",
	CreateKind:   "Create",
}

var (
	OK       = &Msg{Kind: OKKind}
	Stop     = &Msg{Kind: StopKind}
	GetParms = &Msg{Kind: GetParmsKind}
)

type Pipe chan *Msg

type Kind = int

type Msg struct {
	Kind    Kind
	Payload any
	Resp    Pipe
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
