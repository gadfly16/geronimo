package server

import (
	"fmt"
	"net/http"

	mt "github.com/gadfly16/geronimo/messagetypes"
	log "github.com/sirupsen/logrus"
)

func (c *Core) serveHTTP() {
	log.Info("Starting webserver.")

	http.Handle("/", http.FileServer(http.Dir("./gui")))
	http.HandleFunc("/socket", c.wsHandler)

	log.Infoln("Listening on http port: ", c.settings.HTTPAddr)
	err := http.ListenAndServe(c.settings.HTTPAddr, nil)
	c.message <- &Message{
		Type:    mt.WebServerError,
		Payload: err,
	}
}

func (c *Core) wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Received connection request from: ", r.Header.Get("User-Agent"))

	// Upgrade our raw HTTP connection to a websocket based one
	clientID := nextID()
	header := http.Header{}
	header.Set(mt.GeronimoClientID, fmt.Sprint(clientID))
	conn, err := upgrader.Upgrade(w, r, header)
	if err != nil {
		log.Errorln("Error during connection upgradation:", err)
		return
	}

	cl := NewClient(c, conn, clientID)
	c.registerClient <- cl
	go cl.readMessages()
	go cl.writeMessages()
}
