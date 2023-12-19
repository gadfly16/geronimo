package cli

import (
	"flag"
	"fmt"

	mt "github.com/gadfly16/geronimo/messagetypes"
	"github.com/gadfly16/geronimo/server"
	log "github.com/sirupsen/logrus"
)

func init() {
	CommandHandlers["show"] = showCLI
}

func showCLI(s server.Settings) error {
	log.Debug("Running 'show' command.")

	runFlags := flag.NewFlagSet("show", flag.ExitOnError)
	runFlags.Parse(flag.Args()[1:])

	// Connect to server
	conn, clientID, err := connectWSServer(s.WSAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Debugln("Client ID:", clientID)

	msg := server.Message{
		Type:     mt.FullStateRequest,
		ClientID: clientID,
	}

	err = msg.SendWSMessage(conn)
	if err != nil {
		return err
	}

	resp, err := server.ReceiveWSMessage(conn)
	if err != nil {
		return err
	}

	fmt.Println(string(resp.JSPayload))

	return closeServerConnection(conn)
}
