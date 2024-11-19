package core

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func init() {
	nodeMsgHandlers[UserKind] = map[MsgKind]func(Node, *Msg) *Msg{
		UpdateMsgKind:     userUpdateHandler,
		GetParmsMsgKind:   userGetParmsHandler,
		GetCopyMsgKind:    userGetNodeCopyHandler,
		GetDisplayMsgKind: userGetDisplayHandler,
	}
}

type UserParms struct {
	ParmModel
	Admin       bool
	DisplayName string
	Password    []byte
}

type UserNode struct {
	*Head
	Parms *UserParms
}

func (n *UserNode) run() {
	slog.Debug("Running User node.", "name", n.Head.Name)
	for q := range n.In {
		slog.Debug("Message received.", "node", n.path, "kind", q.KindName())
		r := n.Head.handleMsg(n, q)
		q.Answer(r)
		slog.Debug("Message answered.", "node", n.path, "kind", r.KindName())
		if r.Kind == StoppedMsgKind {
			break
		}
	}
	slog.Info("Stopped Root node.")
}

func (t *UserNode) loadBody(h *Head) (n Node, err error) {
	un := &UserNode{
		Head:  h,
		Parms: &UserParms{},
	}
	if err = Db.Where("head_id = ?", h.ID).Order("created_at desc").Take(un.Parms).Error; err != nil {
		return
	}
	return un, nil
}

func (n *UserNode) create(p *Head) (in Pipe, err error) {
	n.OwnerID = n.ID
	n.Head.path = p.path + "/" + n.Name
	n.Parms.Password, err = bcrypt.GenerateFromPassword(n.Parms.Password, 14)
	if err != nil {
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
	go n.run()
	n.Head.initNew()
	slog.Info("Created User node.", "path", n.path)
	return n.Head.In, nil
}

func (n *UserNode) UnmarshalMsg(b io.ReadCloser) (m Msg, err error) {
	m = Msg{
		Payload: UserNode{},
	}
	d := json.NewDecoder(b)
	err = d.Decode(&m)
	return
}

func userGetParmsHandler(ni Node, _ *Msg) *Msg {
	n := ni.(*UserNode)
	return &Msg{
		Kind:    ParmsMsgKind,
		Payload: *n.Parms,
	}
}

func userGetNodeCopyHandler(ni Node, _ *Msg) *Msg {
	ncp := *ni.(*UserNode)
	return &Msg{
		Kind:    ParmsMsgKind,
		Payload: ncp,
	}
}

func userUpdateHandler(ni Node, m *Msg) (r *Msg) {
	n := ni.(*UserNode)
	if m.UserID != n.OwnerID && !m.Admin {
		slog.Debug("unauthorized update request", "path", n.path, "user", m.UserID, "owner", n.OwnerID, "admin", m.Admin)
		return NewErrorMsg(fmt.Errorf("unathorized update request"))
	}
	slog.Debug("user node update", "payload", m.Payload)
	pl := m.Payload.(map[string]any)
	np := &UserParms{
		Admin:       pl["Admin"].(bool),
		DisplayName: pl["Display Name"].(string),
		Password:    n.Parms.Password,
	}
	err := Db.Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Create(np).Error; err != nil {
			return err
		}
		return
	})
	if err != nil {
		return NewErrorMsg(err)
	}
	n.Parms = np
	n.Head.updateGUIs()
	return &OKMsg
}

// func (n *RootNode) setLogLevel() {
// 	LogLevel.Set(slog.Level(n.Parms.LogLevel))
// }

func userGetDisplayHandler(ni Node, _ *Msg) *Msg {
	n := ni.(*UserNode)
	d := n.Head.display()
	d["Parms"] = display{
		"Display Name": n.Parms.DisplayName,
		"Admin":        n.Parms.Admin,
	}
	// slog.Debug("Display data returned by user node", "displayData", d)
	r := &Msg{
		Kind:    DisplayMsgKind,
		Payload: d,
	}
	return r
}
