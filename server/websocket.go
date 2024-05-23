package server

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	WSMsg_Credentials = "Credentials"
	WSMsg_Subscribe   = "Subscribe"
	WSMsq_Unsubscribe = "Unsubscribe"
	WSMsg_Update      = "Update"
)

type guiClient struct {
	id      int64
	otp     string
	conn    *websocket.Conn
	message chan *Message
	user    *Node
}

type socketMessage struct {
	Type   string
	OTP    string
	GUIID  int64
	NodeID uint
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func newGuiClient(conn *websocket.Conn, user *Node) (client *guiClient) {
	client = &guiClient{
		conn:    conn,
		id:      nextID(),
		otp:     generateOTP(),
		message: make(chan *Message),
		user:    user,
	}
	return
}

func websocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorln("Error during connection upgrade:", err)
		return
	}
	gcl := newGuiClient(conn, getRequestUser(c))
	log.Debugf("User: %+v", gcl.user)
	// Send client a new client ID and OTP
	msg := &socketMessage{
		Type:  WSMsg_Credentials,
		GUIID: gcl.id,
		OTP:   gcl.otp,
	}

	err = gcl.sendMessage(msg)
	if err != nil {
		log.Error("Couldn't send client id. Closing connection: ", err)
		conn.Close()
		return
	}

	// Wait for client affirm client credentials
	resp, err := gcl.receiveMessage()
	if err != nil || resp.Type != WSMsg_Credentials {
		log.Error("Error during client id affirmation: ", err)
		conn.Close()
		return
	}

	if resp.GUIID != gcl.id || resp.OTP != gcl.otp {
		log.Error("GUI client credentials aren't matching")
		conn.Close()
		return
	}

	core.guis[gcl.id] = gcl
	go gcl.receiveMessages()
	// go cl.writeMessages()
}

func (gcl *guiClient) receiveMessages() {
eventLoop:
	for {
		msg, err := gcl.receiveMessage()
		if err != nil {
			switch e := err.(type) {
			case *websocket.CloseError:
				if e.Code == 1000 {
					log.Info("Client closed connection gracefully.")
				} else {
					log.Error("Client closed abruptly: ", err)
				}
				break eventLoop
			default:
				log.Errorln("Error during receiving ws message:", err)
				continue eventLoop
			}
		}
		if msg.GUIID != gcl.id || msg.OTP != gcl.otp {
			log.Error("wrong GUI client credentials")
			continue eventLoop
		}
		log.Debugf("Received: %+v", msg)
		switch msg.Type {
		case WSMsg_Subscribe:
			node, err := core.getNode(msg.NodeID, gcl.user)
			if err != nil {
				log.Errorln("subscribing to invalid node:", err)
				continue eventLoop
			}
			core.subscribe(node.ID, gcl)
		case WSMsq_Unsubscribe:
			core.unsubscribe(msg.NodeID, gcl)
		}
	}
	log.Debug("Stopped reading messages for client: ", gcl.id)
}

// func (cl *guiClient) writeMessages() {
// 	for msg := range cl.outgoing {
// 		err := msg.SendWSMessage(cl.conn)
// 		if err != nil {
// 			log.Error("Error writing messages to client (", cl.id, ") :", err)
// 		}
// 	}
// 	log.Debug("Stopped sending messages for client: ", cl.id)
// }

func (gcl *guiClient) sendMessage(msg *socketMessage) (err error) {
	err = gcl.conn.WriteJSON(msg)
	if err != nil {
		return err
	}
	return nil
}

func (gcl *guiClient) receiveMessage() (msg *socketMessage, err error) {
	err = gcl.conn.ReadJSON(&msg)
	if err != nil {
		return nil, err
	}
	return
}

func (gui *guiClient) sendUpdate(nodeid uint) {
	gui.sendMessage(&socketMessage{
		Type:   WSMsg_Update,
		NodeID: nodeid,
	})
}
