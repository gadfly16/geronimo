package tree

import (
	"crypto/rand"
	"log/slog"

	"gorm.io/gorm"
)

var JwtKey []byte

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
	defer close(n.Head.In)
	defer Tree.RemoveNode(n.Head.ID)

	slog.Debug("ROOT node starting up.", "node", n.Head.path, "logLevel", n.Parms.LogLevel)
	for q := range n.In {
		a := handleMsg(n, q)
		if a != nil && a.Kind == M_Stop {
			// Drain unsubscribe messages
			for range len(n.Head.guiSubs) {
				q := <-n.Head.In
				q.AnswerOK()
			}
			q.AnswerMsg(a)
			break
		}
	}
	slog.Info("ROOT stopped.")
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

func (n *RootNode) create(_ []any) (_ *Tag, err error) {
	n.Parms.JwtKey = make([]byte, 14)
	if _, err = rand.Read(n.Parms.JwtKey); err != nil {
		return
	}
	err = Db.Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Create(&n.Head).Error; err != nil {
			return
		}
		n.Parms.HeadID = n.Head.ID
		if err = tx.Create(&n.Parms).Error; err != nil {
			return
		}
		return
	})
	if err != nil {
		return
	}
	n.setLogLevel()
	go n.run()
	n.Head.initNew()
	Tree.Sys.Root = n.Head.Tag
	JwtKey = n.Parms.JwtKey
	return n.Head.Tag, nil
}

func initRootNode(rp *RootParms) (err error) {
	root := newNode(NK_Root).(*RootNode)
	root.Parms = rp
	_, err = root.create(nil)
	return err
}

func rootGetParmsHandler(ni Node, _ *Msg) (r *Msg) {
	n := ni.(*RootNode)
	return &Msg{
		Kind:    M_OK,
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
		Kind:    M_OK,
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
