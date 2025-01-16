package core

import (
	"encoding/json"
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
	Admin    bool
	Email    string
	Password []byte
}

type UserNode struct {
	*Head
	Parms *UserParms
}

func (n *UserNode) run() {
	defer close(n.Head.In)
	defer Tree.RemoveNode(n.Head.ID)

	slog.Debug("Running User node.", "node", n.Head.path)
	for q := range n.In {
		a := n.Head.handleMsg(n, q)
		if a != nil && a.Kind == StoppedMsgKind {
			// Drain unsubscribe messages
			for range len(n.Head.guiSubs) {
				q := <-n.Head.In
				slog.Debug("SINK of User reveived msg.", "node", n.Head.path, "kind", q.KindName())
				q.AnswerOK()
				slog.Debug("SINK of User answered msg.", "node", n.Head.path, "kind", q.KindName())
			}
			q.AnswerMsg(a)
			break
		}
	}

	slog.Info("Stopped User node.", "node", n.Head.path)
}

func (t *UserNode) loadBody(h *Head) (n Node, err error) {
	un := &UserNode{
		Head:  h,
		Parms: &UserParms{},
	}
	if err = Db.Where("head_id = ?", h.ID).Order("created_at desc").Take(un.Parms).Error; err != nil {
		return
	}
	h.Admin = un.Parms.Admin
	return un, nil
}

func (n *UserNode) create(_ []any) (_ *Tag, err error) {
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

	n.Owner = n.Tag
	go n.run()
	n.Head.initNew()
	return n.Head.Tag, nil
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
	var err error
	n := ni.(*UserNode)
	pl := m.Payload.(map[string]any)
	up := &UserParms{
		ParmModel: ParmModel{
			HeadID: n.Head.ID,
		},
		Admin:    pl["Admin"].(bool),
		Email:    pl["Email"].(string),
		Password: n.Parms.Password,
	}
	if pw, ok := pl["Password"]; ok {
		up.Password, err = bcrypt.GenerateFromPassword([]byte(pw.(string)), 14)
		if err != nil {
			return NewErrorMsg(err)
		}
	}
	err = Db.Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Create(up).Error; err != nil {
			return err
		}
		return
	})
	if err != nil {
		return NewErrorMsg(err)
	}
	n.Parms = up
	n.Head.updateGUIs()
	return &OKMsg
}

// func (n *RootNode) setLogLevel() {
// 	LogLevel.Set(slog.Level(n.Parms.LogLevel))
// }

func userGetDisplayHandler(ni Node, _ *Msg) *Msg {
	n := ni.(*UserNode)
	d := n.Head.display()
	d["Parms"] = H{
		"Email":    n.Parms.Email,
		"Admin":    n.Parms.Admin,
		"Password": "",
	}
	// slog.Debug("Display data returned by user node", "displayData", d)
	r := &Msg{
		Kind:    DisplayMsgKind,
		Payload: d,
	}
	return r
}
