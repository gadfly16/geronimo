package server

import (
	"encoding/json"

	mt "github.com/gadfly16/geronimo/messagetypes"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	CLIAgentID = "Geronimo CLI"
)

type MsgID int64

type Message struct {
	ID        MsgID
	Type      string
	ReqID     MsgID
	Payload   interface{} `json:"-"`
	JSPayload json.RawMessage

	RespChan chan *Message `json:"-"`
}

type Settings struct {
	LogLevel       log.Level
	SettingsDbPath string
	HTTPAddr       string
	WSAddr         string
}

type Core struct {
	settings Settings
	accounts []Account
	message  chan *Message
}

func NewCore(s Settings) *Core {
	return &Core{settings: s, message: make(chan *Message)}
}

func (c *Core) Run() error {
	log.Info("Running core.")

	s := newWebServer(c)
	go s.Run()

	// Load state
	db, err := gorm.Open(sqlite.Open(c.settings.SettingsDbPath), &gorm.Config{})
	if err != nil {
		return err
	}
	db.Find(&c.accounts)

	for {
		req := <-c.message
		resp := &Message{Type: mt.Done, ReqID: req.ID}
		switch req.Type {
		case mt.CreateAccount:
			acc := req.Payload.(Account)
			err = acc.Save(db)
			if err != nil {
				resp.Type = mt.Error
			}
		case mt.FullStateRequest:
			resp.Type = mt.FullState
			resp.JSPayload, err = json.Marshal(c.accounts)
			if err != nil {
				resp.Type = mt.Error
			}
		case mt.WebServerError:
			return req.Payload.(error)
		default:
			log.Errorln("Unknown core request type.")
		}
		if resp.Type == mt.Error {
			resp.Payload = err.Error()
		}
		req.RespChan <- resp
	}
}
