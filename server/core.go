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
	LogLevel      string `short:"l" long:"log-level" description:"logging level" default:"debug" choice:"info" choice:"debug"`
	WorkDir       string `short:"w" long:"work-dir" description:"work directory" default-mask:"$HOME/.config/Geronimo"`
	HTTPAddr      string `short:"A" long:"http-address" description:"http address" default:"localhost:8088"`
	UserEmail     string `short:"u" long:"user-email" description:"user email address"`
	UserPassword  string `short:"p" long:"user-password" description:"user password"`
	WSAddr        string
	DBPath        string
	JWTKeyPath    string
	DBKeyPath     string
	CLICookiePath string
}

type Core struct {
	settings         Settings
	root             *Node
	nodes            map[uint]*Node
	clients          map[int64]*Client
	message          chan *Message
	db               *gorm.DB
	registerClient   chan *Client
	unregisterClient chan *Client
	jwtKey           []byte
	dbKey            []byte
}

func (s *Settings) Init() {
	logLevel, _ := log.ParseLevel(s.LogLevel)
	log.SetLevel(logLevel)

	if s.WorkDir == "" {
		s.WorkDir = os.Getenv("HOME") + "/.config/Geronimo"
	}
	s.DBPath = s.WorkDir + "/" + NameStateDB
	s.DBKeyPath = s.WorkDir + "/" + NameDBKey
	s.JWTKeyPath = s.WorkDir + "/" + NameJWTKey
	s.CLICookiePath = s.WorkDir + "/" + NameCLICookie
	s.WSAddr = "ws://" + s.HTTPAddr + "/socket"
}

func Init(s Settings) (err error) {
	if err = createDB(s); err != nil {
		return
	}

	if err = createSecret(s.JWTKeyPath); err != nil {
		return
	}

	if err = createSecret(s.DBKeyPath); err != nil {
		return
	}
	return
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
		root:             &Node{DetailType: NodeRoot},
		nodes:            map[uint]*Node{},
		clients:          map[int64]*Client{},
		message:          make(chan *Message),
		registerClient:   make(chan *Client),
		unregisterClient: make(chan *Client),
	}
	core.nodes[0] = core.root
	// Connect to db
	core.db, err = gorm.Open(sqlite.Open(s.DBPath), &gorm.Config{})
	if err != nil {
		return
	}

	if err = core.loadChildren(core.root); err != nil {
		return
	}

	// Load secrets
	if core.jwtKey, err = os.ReadFile(s.JWTKeyPath); err != nil {
		return
	}
	if core.dbKey, err = os.ReadFile(s.DBKeyPath); err != nil {
		return
	}
	return
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
