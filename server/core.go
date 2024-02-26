package server

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	CLIAgentID = "Geronimo CLI"

	NameStateDB   = "state.db"
	NameJWTKey    = "jwt-key"
	NameDBKey     = "db-key"
	NameCLICookie = "cli-cookie"

	ExiprationMins = 60

	GeronimoClientID = "Geronimo-Client-ID"
)

type ErrorPayload struct {
	Status int
	Error  string
}

type Settings struct {
	LogLevel      log.Level
	WorkDir       string
	HTTPAddr      string
	WSAddr        string
	UserEmail     string
	UserPassword  string
	DBPath        string
	JWTKeyPath    string
	DBKeyPath     string
	CLICookiePath string
}

type Core struct {
	settings         Settings
	users            []*User
	userMap          map[uint]*User
	accountMap       map[uint]*Account
	brokerMap        map[uint]*Broker
	clients          map[int64]*Client
	message          chan *Message
	db               *gorm.DB
	registerClient   chan *Client
	unregisterClient chan *Client
	jwtKey           []byte
	dbKey            []byte
}

func Init(s Settings) (err error) {
	if err = createDB(s); err != nil {
		return
	}
	if err = createSecret(s.JWTKeyPath); err != nil {
		return
	}
	return createSecret(s.DBKeyPath)
}

func Serve(s Settings) error {
	c, err := newCore(s)
	if err != nil {
		return err
	}

	go c.serveHTTP()

	return c.Run()
}

func newCore(s Settings) (core *Core, err error) {
	core = &Core{
		settings:         s,
		users:            []*User{},
		userMap:          map[uint]*User{},
		accountMap:       map[uint]*Account{},
		brokerMap:        map[uint]*Broker{},
		clients:          map[int64]*Client{},
		message:          make(chan *Message),
		registerClient:   make(chan *Client),
		unregisterClient: make(chan *Client),
	}

	// Connect to db
	core.db, err = gorm.Open(sqlite.Open(s.DBPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Load models
	res := core.db.Preload("Accounts.Brokers").Find(&core.users)
	if res.Error != nil {
		return nil, err
	}

	// Fill out model maps and initialize states
	for _, u := range core.users {
		core.userMap[u.ID] = u
		for _, a := range u.Accounts {
			core.accountMap[a.ID] = a
			for _, b := range a.Brokers {
				core.brokerMap[b.ID] = b
			}
		}
	}

	// Load secrets
	if core.jwtKey, err = os.ReadFile(s.JWTKeyPath); err != nil {
		return nil, err
	}
	if core.dbKey, err = os.ReadFile(s.DBKeyPath); err != nil {
		return nil, err
	}
	return core, nil
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
			if mh, ok := messageHandlers[req.Type]; ok {
				resp := mh(c, req)
				req.RespChan <- resp
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

// func (core *Core) initAccount(acc *Account) (err error) {
// 	acc.ueApiPublicKey, err = decryptString(core.dbKey, acc.Name, acc.ApiPublicKey)
// 	if err != nil {
// 		return err
// 	}
// 	acc.ueApiPrivateKey, err = decryptString(core.dbKey, acc.Name, acc.ApiPrivateKey)
// 	if err != nil {
// 		return err
// 	}
// 	acc.ApiPublicKey = ""
// 	acc.ApiPrivateKey = ""
// 	return
// }
