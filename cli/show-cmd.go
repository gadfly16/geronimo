package cli

import (
	"flag"
	"fmt"

	mt "github.com/gadfly16/geronimo/messagetypes"
	"github.com/gadfly16/geronimo/server"
	log "github.com/sirupsen/logrus"
)

func init() {
	Commands["show"] = showCLI
}

func showCLI(s server.Settings) error {
	log.Debug("Running 'show' command.")

	runFlags := flag.NewFlagSet("show", flag.ExitOnError)
	runFlags.Parse(flag.Args()[1:])

	// Connect to server
	conn, err := connectWSServer(s.WSAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	msg := server.Message{
		Type: mt.FullStateRequest,
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

	err = server.CloseServerConnection(conn)
	if err != nil {
		return err
	}
	return nil
}
