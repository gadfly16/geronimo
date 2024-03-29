package server

import (
	"encoding/json"
	"errors"

	"github.com/gin-gonic/gin"
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

var upgrader = websocket.Upgrader{}

func (core *Core) wsHandler(c *gin.Context) {
	log.Debug("Received connection request from: ", c.Request.Header.Get("User-Agent"))

	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorln("Error during connection upgradation:", err)
		return
	}

	// Send client a new client ID
	clid := nextID()
	msg := &Message{
		Type:    MessageClientID,
		Payload: clid,
	}
	log.Debug("Sending new client ID: ", clid)
	err = msg.SendWSMessage(conn)
	if err != nil {
		log.Error("Couldn't send client id. Closing connection: ", err)
		conn.Close()
		return
	}

	// Wait for client affirm client id
	resp, err := ReceiveWSMessage(conn)
	if err != nil || resp.Type != MessageClientID {
		log.Error("Error during client id affirmation. Closing connection.")
		conn.Close()
		return
	}

	cl := NewClient(core, conn, clid)
	core.registerClient <- cl
	go cl.readMessages()
	go cl.writeMessages()
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
					log.Error("Client closed abruptly: ", err)
				}
				close(cl.outgoing)
				cl.core.unregisterClient <- cl
				break eventLoop
			default:
				log.Errorln("Error during receiving ws message:", err)
				continue eventLoop
			}
		}
		log.Debugln("Websocket server received:", msg.Type)
		cl.core.message <- msg
	}
	log.Debug("Stopped reading messages for client: ", cl.id)
}

func (cl *Client) writeMessages() {
	for msg := range cl.outgoing {
		err := msg.SendWSMessage(cl.conn)
		if err != nil {
			log.Error("Error writing messages to client (", cl.id, ") :", err)
		}
	}
	log.Debug("Stopped sending messages for client: ", cl.id)
}

func (msg *Message) SendWSMessage(conn *websocket.Conn) (err error) {
	// Some messages are already in JSON because of deep copying structures
	if msg.Type != MessageState {
		msg.JSPayload, err = json.Marshal(msg.Payload)
		if err != nil {
			return err
		}
	}

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
	case MessageError:
		var errMsg string
		err = json.Unmarshal(msg.JSPayload, &errMsg)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(errMsg)
	case MessageClientID:
		var clid int64
		err = json.Unmarshal(msg.JSPayload, &clid)
		if err != nil {
			return nil, err
		}
		msg.Payload = clid
	case MessageCreateAccount, MessageAccount:
		var acc Account
		err = json.Unmarshal(msg.JSPayload, &acc)
		if err != nil {
			return nil, err
		}
		msg.Payload = acc
	case MessageState:
		var accs []Account
		err = json.Unmarshal(msg.JSPayload, &accs)
		if err != nil {
			return nil, err
		}
		msg.Payload = accs
	}
	return
}
