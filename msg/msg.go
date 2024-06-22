package msg

const (
	OKKind = iota
	ErrorKind
	UpdateKind
)

type Pipe chan *Msg

type Kind = int

type Msg struct {
	Kind    Kind
	Payload map[string]any
	Resp    Pipe
	Error   error
}
