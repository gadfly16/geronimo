package msg

const (
	OKKind = iota
	ErrorKind
	StopRootKind
	UpdateKind
	ParmsKind
	GetParmsKind
	CreateKind
)

var kindNames = map[Kind]string{
	OKKind:       "OK",
	ErrorKind:    "Error",
	StopRootKind: "StopTree",
	UpdateKind:   "Update",
	ParmsKind:    "Parms",
	GetParmsKind: "GetParms",
	CreateKind:   "Create",
}

var OK = &Msg{
	Kind: OKKind,
}

type Pipe chan *Msg

type Kind = int

type Msg struct {
	Kind    Kind
	Payload any
	Resp    Pipe
}

func (m *Msg) Ask(out Pipe) *Msg {
	rin := make(Pipe)
	m.Resp = rin
	out <- m
	return <-rin
}

func (m *Msg) Answer(q *Msg) {
	q.Resp <- m
}

func NewError(err error) *Msg {
	return &Msg{
		Kind:    ErrorKind,
		Payload: err.Error(),
	}
}

func (m *Msg) Error() string {
	return m.Payload.(string)
}

func (m *Msg) KindName() string {
	return kindNames[m.Kind]
}
