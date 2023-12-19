package cli

import (
	"fmt"
	"net/http"
	"strconv"
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
	conn, resp, err := websocket.DefaultDialer.Dial(WSAddr, reqHeader)
	if err != nil {
		return nil, 0, err
	}

	clientID, err := strconv.ParseInt(resp.Header.Get(mt.GeronimoClientID), 10, 64)
	if err != nil {
		return nil, 0, err
	}
	return conn, clientID, nil
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
