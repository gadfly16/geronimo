package msg

const (
	OKKind = iota
	ErrorKind
	UpdateKind
	ParmsKind
	GetParmsKind
)

type Pipe chan *Msg

type Kind = int

type Msg struct {
	Kind    Kind
	Payload any
	Resp    Pipe
	Error   error
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
