package node

import (
	"crypto/rand"
	"log/slog"

	"github.com/gadfly16/geronimo/msg"
	"gorm.io/gorm"
)

var JwtKey []byte

func init() {
	nodeMsgHandlers[RootKind] = map[msg.Kind]func(Node, *msg.Msg) *msg.Msg{
		// msg.UpdateKind:   rootUpdateHandler,
		msg.GetParmsKind: rootGetParmsHandler,
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
		slog.Debug("Message received.", "node", n.Name, "kind", q.KindName())
		r := n.Head.handleMsg(n, q)
		r.Answer(q)
		slog.Debug("Message answered.", "node", n.Name, "kind", r.KindName())
		if r.Kind == msg.StoppedKind {
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

func (n *RootNode) create(pp string) (in msg.Pipe, err error) {
	n.Head.path = pp + "/" + n.Head.Name
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
	n.Head.register()
	Tree.Root = n.Head.In
	JwtKey = n.Parms.JwtKey
	slog.Info("Created Root node.", "path", n.Head.path)
	return n.Head.In, nil
}

func InitRootNode(rp *RootParms) (err error) {
	root := &RootNode{
		Head: &Head{
			Name: "Root",
			Kind: RootKind,
		},
		Parms: rp,
	}
	_, err = root.create("")
	return err
}

func rootGetParmsHandler(ni Node, m *msg.Msg) (r *msg.Msg) {
	n := ni.(*RootNode)
	return &msg.Msg{
		Kind:    msg.ParmsKind,
		Payload: *n.Parms,
	}
}

// func rootUpdateHandler(ni Node, m *msg.Msg) (r *msg.Msg) {
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

func (n *RootNode) setLogLevel() {
	LogLevel.Set(slog.Level(n.Parms.LogLevel))
}
