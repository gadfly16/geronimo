package server

import (
	"encoding/json"
	"errors"
	"net/http"

	mt "github.com/gadfly16/geronimo/messagetypes"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{}

type webserver struct {
	core *Core
}

func newWebServer(core *Core) *webserver {
	return &webserver{core}
}

func (s *webserver) Run() {
	log.Info("Running webserver.")

	http.Handle("/", http.FileServer(http.Dir("./gui")))
	http.HandleFunc("/socket", s.socketHandler)

	log.Infoln("Listening on http port: ", s.core.settings.HTTPAddr)
	err := http.ListenAndServe(s.core.settings.HTTPAddr, nil)
	s.core.message <- &Message{Type: mt.WebServerError, Payload: err}
}

func (s *webserver) socketHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Received connection request from: ", r.Header["User-Agent"][0])

	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorln("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()

	fromCore := make(chan *Message)

eventLoop:
	for {
		msg, err := ReceiveWSMessage(conn)
		if err != nil {
			switch e := err.(type) {
			case *websocket.CloseError:
				if e.Code == 1000 {
					log.Info("Client closed connection gracefully.")
				} else {
					log.Errorln("Client closed abruptly: ", err)
				}
				break eventLoop
			default:
				log.Errorln("Error during receiving ws message:", err)
				continue eventLoop
			}
		}
		log.Debugln("Websocket server received:", msg.Type)

		msg.RespChan = fromCore
		s.core.message <- msg
		resp := <-fromCore

		log.Debugln("Websocket server sending:", resp.Type)
		err = resp.SendWSMessage(conn)
		if err != nil {
			log.Errorln("Error during sending ws message: ", err)
		}
	}
}

func (msg *Message) SendWSMessage(conn *websocket.Conn) (err error) {
	// Some messages are alreadi in JSON for deep copy
	switch msg.Type {
	case mt.FullState:
	default:
		msg.JSPayload, err = json.Marshal(msg.Payload)
		if err != nil {
			return err
		}
	}

	msg.ID = nextMsgID()
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
	case mt.CreateAccount:
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

func CloseServerConnection(conn *websocket.Conn) error {
	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return err
	}
	return nil
}
