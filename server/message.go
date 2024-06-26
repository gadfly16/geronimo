package server

const (
	MessageOK                   = "OK"
	MessageError                = "Error"
	MessageGUIClientCredentials = "GUIClientCredentials"
	MessageAccount              = "Account"
	MessageCommandResponse      = "CommandResponse"
	MessageGetTree              = "GetState"
	MessageTree                 = "State"
	MessageWebServerError       = "WebServerError"
	MessageAuthenticateUser     = "AuthenticateUser"
	MessageUser                 = "UserWithSecret"
	MessageCreateUser           = "CreateUser"
	MessageCreate               = "Create"
	MessageUpdate               = "Update"
	MessageGetDisplay           = "GetDetail"
	MessageDisplay              = "Display"
)

type Message struct {
	Type    string
	User    *Node
	Path    string
	Payload interface{}

	respChan chan *Message
}

func (core *Core) send(msgType string, payload interface{}) *Message {
	msg := &Message{
		Type:     msgType,
		Payload:  payload,
		respChan: make(chan *Message),
	}
	core.message <- msg
	return <-msg.respChan
}

func (msg *Message) toCore() *Message {
	msg.respChan = make(chan *Message)
	core.message <- msg
	return <-msg.respChan
}

func errorMessage(status int, errMsg string) *Message {
	return &Message{
		Type:    MessageError,
		Payload: &ErrorPayload{status, errMsg},
	}
}

func (msg *Message) extractError() (int, *APIError) {
	e := msg.Payload.(*ErrorPayload)
	return e.Status, &APIError{Error: e.Error}
}

// func (c *Core) broadcastMessage(msg *Message) {
// 	for _, cl := range c.clients {
// 		cl.outgoing <- msg
// 	}
// }
