package cli

import (
	"errors"
	"fmt"
	"net/http"
	"syscall"

	mt "github.com/gadfly16/geronimo/messagetypes"
	"github.com/gadfly16/geronimo/server"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
)

type commandHandler func(server.Settings) error

var CommandHandlers = make(map[string]commandHandler)

func getTerminalString(prompt string) string {
	fmt.Print(prompt)
	pw, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("Couldn't get password.")
	}
	fmt.Println()
	return string(pw)
}

func connectWSServer(WSAddr string) (*websocket.Conn, int64, error) {
	reqHeader := http.Header{"User-Agent": []string{server.CLIAgentID}}
	conn, _, err := websocket.DefaultDialer.Dial(WSAddr, reqHeader)
	if err != nil {
		return nil, 0, err
	}

	msg, err := server.ReceiveWSMessage(conn)
	if err != nil {
		return nil, 0, err
	}
	if msg.Type != mt.ClientID {
		return nil, 0, errors.New("didn't received client id message, but: " + msg.Type)
	}
	clid := msg.Payload.(int64)
	msg.ClientID = clid
	err = msg.SendWSMessage(conn)
	if err != nil {
		return nil, 0, err
	}

	return conn, clid, nil
}

func waitForResponse(conn *websocket.Conn, msg *server.Message) (resp *server.Message, err error) {
	for {
		resp, err = server.ReceiveWSMessage(conn)
		if err != nil {
			return
		}
		if resp.ClientID == msg.ClientID && resp.ReqID == msg.ID {
			break
		}
	}
	return
}

func closeServerConnection(conn *websocket.Conn) error {
	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return err
	}
	return nil
}
