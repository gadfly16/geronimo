package core

import (
	"crypto/rand"
	"log/slog"

	"gorm.io/gorm"
)

var JwtKey []byte

func init() {
	nodeMsgHandlers[RootKind] = map[MsgKind]func(Node, *Msg) *Msg{
		// msg.UpdateKind:   rootUpdateHandler,
		GetParmsMsgKind:   rootGetParmsHandler,
		GetDisplayMsgKind: rootGetDisplayHandler,
	}
}

type RootParms struct {
	ParmModel
	LogLevel int
	HTTPAddr string
	DbKey    string
	JwtKey   []byte
}

type RootNode struct {
	*Head
	Parms *RootParms
}

var LogLevel = new(slog.LevelVar)
var LogLevelNames = map[string]slog.Level{
	"info":  slog.LevelInfo,
	"debug": slog.LevelDebug,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func (n *RootNode) run() {
	slog.Info("Running Root node.", "name", n.Head.Name, "logLevel", n.Parms.LogLevel)
	for q := range n.In {
		a := n.Head.handleMsg(n, q)
		if a != nil && a.Kind == StoppedMsgKind {
			break
		}
	}
	slog.Info("Stopped Root node.")
}

func (nt *RootNode) loadBody(h *Head) (n Node, err error) {
	rn := &RootNode{
		Head:  h,
		Parms: &RootParms{},
	}
	if err = Db.Where("head_id = ?", h.ID).Order("created_at desc").Take(rn.Parms).Error; err != nil {
		return
	}
	rn.setLogLevel()
	JwtKey = rn.Parms.JwtKey
	return rn, nil
}

func (n *RootNode) create() (in Pipe, err error) {
	n.Parms.JwtKey = make([]byte, 14)
	if _, err = rand.Read(n.Parms.JwtKey); err != nil {
		return
	}
	err = Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&n.Head).Error; err != nil {
			return err
		}
		n.Parms.HeadID = n.Head.ID
		if err := tx.Create(&n.Parms).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return
	}
	n.setLogLevel()
	go n.run()
	n.Head.initNew()
	Tree.Sys.Root = n.Head.In
	JwtKey = n.Parms.JwtKey
	slog.Info("Created Root node.", "path", n.Head.path)
	return n.Head.In, nil
}

func initRootNode(rp *RootParms) (err error) {
	root := NewNodeKind(RootKind).(*RootNode)
	root.Parms = rp
	_, err = root.create()
	return err
}

func rootGetParmsHandler(ni Node, m *Msg) (r *Msg) {
	n := ni.(*RootNode)
	return &Msg{
		Kind:    ParmsMsgKind,
		Payload: *n.Parms,
	}
}

func rootGetDisplayHandler(ni Node, _ *Msg) *Msg {
	n := ni.(*RootNode)
	d := n.Head.display()
	// d["Parms"] = display{
	// 	"Display Name": n.Parms.DisplayName,
	// }
	r := &Msg{
		Kind:    DisplayMsgKind,
		Payload: d,
	}
	return r
}

// func rootUpdateHandler(ni Node, m *Msg) (r *Msg) {
// if v, ok := m.Payload["parms"]; ok {
// 	if n.Parms, ok = v.(*RootParms); !ok {
// 		return fmt.Errorf("update failed: wrong parms type %T", n.Parms)
// 	}
// }
// Db.Transaction(func(tx *gorm.DB) error {
// 	return nil
// })
// 	return
// }

func (n *RootNode) getDisplay() (d H) {
	d = H{
		"Parms": n.Parms,
	}
	return
}

func (n *RootNode) setLogLevel() {
	LogLevel.Set(slog.Level(n.Parms.LogLevel))
}
