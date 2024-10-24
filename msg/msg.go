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
	GetDisplayKind
	DisplayKind
	SubscribeKind
	UnsubscribeKind
	UpdatedKind
)

var KindNames = map[Kind]string{
	OKKind:          "OK",
	ErrorKind:       "Error",
	StopKind:        "Stop",
	StoppedKind:     "Stopped",
	UpdateKind:      "Update",
	ParmsKind:       "Parms",
	GetParmsKind:    "GetParms",
	CreateKind:      "Create",
	AuthUserKind:    "AuthUser",
	GetTreeKind:     "GetTree",
	TreeKind:        "Tree",
	GetCopyKind:     "GetCopy",
	GetDisplayKind:  "GetDisplay",
	DisplayKind:     "Display",
	SubscribeKind:   "Subscribe",
	UnsubscribeKind: "Unsubscribe",
	UpdatedKind:     "Updated",
}

var (
	OK         = Msg{Kind: OKKind}
	Stop       = Msg{Kind: StopKind}
	Stopped    = Msg{Kind: StoppedKind}
	GetParms   = Msg{Kind: GetParmsKind}
	GetCopy    = Msg{Kind: GetCopyKind}
	GetTree    = Msg{Kind: GetTreeKind}
	GetDisplay = Msg{Kind: GetDisplayKind}
	Updated    = Msg{Kind: UpdatedKind}
)

type Pipe chan *Msg

type Kind = int

type Msg struct {
	Kind    Kind
	Payload any
	Resp    Pipe

	UserID int
	Admin  bool
}

func (q *Msg) Answer(m *Msg) {
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
	return KindNames[m.Kind]
}

func (t Pipe) Ask(m Msg) Msg {
	m.Resp = make(Pipe)
	t <- &m
	return *<-m.Resp
}
