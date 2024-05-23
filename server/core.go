package server

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var core *Core

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
	LogLevel      string
	WorkDir       string
	HTTPAddr      string
	WSAddr        string
	DBPath        string
	JWTKeyPath    string
	DBKeyPath     string
	CLICookiePath string
}

type Core struct {
	settings      Settings
	root          *Node
	nodes         map[uint]*Node
	guis          map[int64]*guiClient
	subscriptions map[uint]map[*guiClient]bool
	message       chan *Message
	db            *gorm.DB
	jwtKey        []byte
	dbKey         []byte
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

func Serve(s Settings) (err error) {
	core, err = newCore(s)
	if err != nil {
		return err
	}

	go core.serveHTTP()

	return runCore()
}

func newCore(s Settings) (core *Core, err error) {
	core = &Core{
		settings:      s,
		root:          &Node{DetailType: NodeRoot},
		nodes:         map[uint]*Node{},
		guis:          map[int64]*guiClient{},
		subscriptions: map[uint]map[*guiClient]bool{},
		message:       make(chan *Message),
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

func (core *Core) subscribe(nodeid uint, gui *guiClient) {
	if _, ok := core.subscriptions[nodeid]; !ok {
		core.subscriptions[nodeid] = make(map[*guiClient]bool)
	}
	core.subscriptions[nodeid][gui] = true
}

func (core *Core) unsubscribe(nodeid uint, gui *guiClient) {
	delete(core.subscriptions[nodeid], gui)
	if len(core.subscriptions[nodeid]) == 0 {
		delete(core.subscriptions, nodeid)
	}
}

func (core *Core) sendUpdates(nodeid uint) {
	if guis, ok := core.subscriptions[nodeid]; ok {
		for gui := range guis {
			gui.sendUpdate(nodeid)
		}
	}
}

func runCore() (err error) {
	log.Info("Starting core.")
	for {
		select {
		case req := <-core.message:
			if mh, ok := messageHandlers[req.Type]; ok {
				resp := mh(req)
				req.respChan <- resp
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
