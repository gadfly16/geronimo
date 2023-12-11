package cli

import (
	"fmt"
	"net/http"
	"syscall"

	"github.com/gadfly16/geronimo/server"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
)

var Commands = make(map[string]func(server.Settings) error)

func getTerminalString(prompt string) string {
	fmt.Print(prompt)
	pw, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("Couldn't get password.")
	}
	fmt.Println()
	return string(pw)
}

func connectWSServer(WSAddr string) (*websocket.Conn, error) {
	reqHeader := http.Header{"User-Agent": []string{server.CLIAgentID}}
	conn, _, err := websocket.DefaultDialer.Dial(WSAddr, reqHeader)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
