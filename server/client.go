package server

import (
	"encoding/json"
	"errors"

	mt "github.com/gadfly16/geronimo/messagetypes"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	id       int64
	core     *Core
	conn     *websocket.Conn
	message  chan *Message
	incoming chan *Message
	outgoing chan *Message
}

func NewClient(c *Core, conn *websocket.Conn, id int64) *Client {
	client := &Client{core: c, conn: conn, id: id}
	client.message = make(chan *Message)
	client.incoming = make(chan *Message)
	client.outgoing = make(chan *Message)
	return client
}

func (cl *Client) directMessage(msg *Message) error {
	return msg.SendWSMessage(cl.conn)
}

func (cl *Client) readMessages() {
eventLoop:
	for {
		msg, err := ReceiveWSMessage(cl.conn)
		if err != nil {
			switch e := err.(type) {
			case *websocket.CloseError:
				if e.Code == 1000 {
					log.Info("Client closed connection gracefully.")
				} else {
					log.Errorln("Client closed abruptly: ", err)
				}
				cl.core.unregisterClient <- cl
			default:
				log.Errorln("Error during receiving ws message:", err)
				continue eventLoop
			}
		}
		log.Debugln("Websocket server received:", msg.Type)
		cl.core.message <- msg
	}
}

func (cl *Client) writeMessages() {

}

func (msg *Message) SendWSMessage(conn *websocket.Conn) (err error) {
	// Some messages are already in JSON because of deep copying structures
	if msg.Type != mt.FullState {
		msg.JSPayload, err = json.Marshal(msg.Payload)
		if err != nil {
			return err
		}
	}

	msg.ID = nextID()
	err = conn.WriteJSON(msg)
	if err != nil {
		return err
	}
	return nil
}

func ReceiveWSMessage(conn *websocket.Conn) (msg *Message, err error) {
	err = conn.ReadJSON(&msg)
	if err != nil {
		return nil, err
	}

	switch msg.Type {
	case mt.Error:
		var errMsg string
		err = json.Unmarshal(msg.JSPayload, &errMsg)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(errMsg)
	case mt.CreateAccount, mt.NewAccount:
		var acc Account
		err = json.Unmarshal(msg.JSPayload, &acc)
		if err != nil {
			return nil, err
		}
		msg.Payload = acc
	case mt.FullState:
		var accs []Account
		err = json.Unmarshal(msg.JSPayload, &accs)
		if err != nil {
			return nil, err
		}
		msg.Payload = accs
	}
	return
}
