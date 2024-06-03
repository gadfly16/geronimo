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
	err = initCore(s)
	if err != nil {
		return err
	}

	go core.serveHTTP()

	return runCore()
}

func initCore(s Settings) (err error) {
	core = &Core{
		settings:      s,
		root:          &Node{DetailType: NodeRoot, Detail: &Root{}},
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

	// Load secrets
	if core.jwtKey, err = os.ReadFile(s.JWTKeyPath); err != nil {
		return
	}
	if core.dbKey, err = os.ReadFile(s.DBKeyPath); err != nil {
		return
	}

	if err = core.root.loadChildren(); err != nil {
		return
	}
	return core.root.run()
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
	log.Info("Core starts handling messages.")
	for req := range core.message {
		if mh, ok := messageHandlers[req.Type]; ok {
			resp := mh(req)
			req.respChan <- resp
		} else {
			log.Errorln("Received unknown message type.")
		}
	}
	return
}
