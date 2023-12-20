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

type Message struct {
	ID, ReqID, ClientID int64
	Type                string
	Payload             interface{} `json:"-"`
	JSPayload           json.RawMessage

	RespChan chan *Message `json:"-"`
}

type Settings struct {
	LogLevel       log.Level
	SettingsDbPath string
	HTTPAddr       string
	WSAddr         string
}

type Core struct {
	settings         Settings
	accounts         []Account
	clients          map[int64]*Client
	message          chan *Message
	db               *gorm.DB
	registerClient   chan *Client
	unregisterClient chan *Client
}

type messageHandler func(*Core, *Message) error

var messageHandlers = map[string]messageHandler{
	mt.CreateAccount:    createAccountHandler,
	mt.FullStateRequest: fullStateRequestHandler,
}

func Serve(s Settings) error {
	c, err := newCore(s)
	if err != nil {
		return err
	}

	go c.serveHTTP()

	return c.Run()
}

func newCore(s Settings) (c *Core, err error) {
	c = &Core{
		settings:         s,
		clients:          map[int64]*Client{},
		message:          make(chan *Message),
		registerClient:   make(chan *Client),
		unregisterClient: make(chan *Client),
	}
	// Load state
	c.db, err = gorm.Open(sqlite.Open(c.settings.SettingsDbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	res := c.db.Find(&c.accounts)
	if res.Error != nil {
		return nil, err
	}
	return c, nil
}

func (c *Core) Run() (err error) {
	log.Info("Starting core.")
	for {
		select {
		case cl := <-c.registerClient:
			c.clients[cl.id] = cl
			log.Infoln("Registered client:", cl.id)
		case cl := <-c.unregisterClient:
			if _, ok := c.clients[cl.id]; ok {
				delete(c.clients, cl.id)
				cl.conn.Close()
				log.Info("Unregistered client: ", cl.id)
			} else {
				log.Error("Can't unregister unregistered client: ", cl.id)
			}
		case req := <-c.message:
			if req.ClientID == 0 {
				log.Error("Received message with no client id.")
				continue
			}
			if mh, ok := messageHandlers[req.Type]; ok {
				err = mh(c, req)
				if err != nil {
					resp := req.prepareResponse()
					resp.Type = mt.Error
					resp.Payload = err
					c.clients[req.ClientID].outgoing <- resp
				}
			} else {
				log.Errorln("Reveived unknown message type.")
			}
			// resp := &Message{Type: mt.Done, ReqID: req.ID}
			// switch req.Type {
			// case mt.CreateAccount:
			// 	acc := req.Payload.(Account)
			// 	err = acc.Save(c.db)
			// 	if err != nil {
			// 		resp.Type = mt.Error
			// 	}
			// case mt.FullStateRequest:
			// 	resp.Type = mt.FullState
			// 	resp.JSPayload, err = json.Marshal(c.accounts)
			// 	if err != nil {
			// 		resp.Type = mt.Error
			// 	}
			// case mt.WebServerError:
			// 	return req.Payload.(error)
			// default:
			// 	log.Errorln("Unknown core request type.")
			// }
			// if resp.Type == mt.Error {
			// 	resp.Payload = err.Error()
			// }
			// req.RespChan <- resp
		}
	}
}

func (msg *Message) prepareResponse() *Message {
	return &Message{ReqID: msg.ID, ClientID: msg.ClientID}
}

func (c *Core) broadcastMessage(msg *Message) {
	for _, cl := range c.clients {
		cl.outgoing <- msg
	}
}

func createAccountHandler(c *Core, msg *Message) (err error) {
	resp := msg.prepareResponse()
	resp.Type = mt.NewAccount

	acc := msg.Payload.(Account)
	err = acc.Save(c.db)
	if err != nil {
		return err
	}
	c.accounts = append(c.accounts, acc)
	resp.Payload = acc
	c.broadcastMessage(resp)
	log.Info("Created new account: ", acc.Name)
	return nil
}

func fullStateRequestHandler(c *Core, msg *Message) (err error) {
	log.Debugln("Handling full state request.")
	resp := msg.prepareResponse()
	resp.Type = mt.FullState

	resp.JSPayload, err = json.Marshal(c.accounts)
	if err != nil {
		return err
	}
	c.clients[msg.ClientID].outgoing <- resp
	return nil
}
