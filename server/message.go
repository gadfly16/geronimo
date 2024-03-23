package server

import "encoding/json"

const (
	MessageOK               = "OK"
	MessageError            = "Error"
	MessageClientID         = "ClientID"
	MessageCreateAccount    = "CreateAccount"
	MessageAccount          = "Account"
	MessageCommandResponse  = "CommandResponse"
	MessageGetState         = "GetState"
	MessageState            = "State"
	MessageWebServerError   = "WebServerError"
	MessageAuthenticateUser = "AuthenticateUser"
	MessageUser             = "UserWithSecret"
	MessageCreateUser       = "CreateUser"
)

type Message struct {
	Type      string
	Payload   interface{} `json:"-"`
	JSPayload json.RawMessage

	RespChan chan *Message `json:"-"`
}

func (core *Core) send(msgType string, payload interface{}) *Message {
	msg := &Message{
		Type:     msgType,
		Payload:  payload,
		RespChan: make(chan *Message),
	}
	core.message <- msg
	return <-msg.RespChan
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

func (c *Core) broadcastMessage(msg *Message) {
	for _, cl := range c.clients {
		cl.outgoing <- msg
	}
}
