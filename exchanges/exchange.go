package exchanges

const (
	MsgOK = iota
	MsgError
	MsgSubscribeBalance
	MsgBalanceUpdate
)

type MsgPayload = map[string]interface{}

type Msg struct {
	Kind    int
	Payload MsgPayload
}

type Conn struct {
	In  chan *Msg
	Out chan *Msg
}

func NewConnection() (conn *Conn) {
	return &Conn{
		In:  make(chan *Msg),
		Out: make(chan *Msg),
	}
}

func ErrorMsg(errorMsg string) (resp *Msg) {
	return &Msg{
		Kind: MsgError,
		Payload: MsgPayload{
			"Error": errorMsg,
		},
	}
}
