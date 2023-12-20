package server

import (
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
