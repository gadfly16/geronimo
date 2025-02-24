package tree

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/coder/websocket"
)

// Payload prototypes
var payloadProtos []any = []any{
	M_Get_Tree:     nil,
	M_Get_Display:  nil,
	M_Create:       []any{0, "", H{}},
	M_Rename:       []any{"", ""},
	M_Delete:       []any{""},
	M_Update_Parms: H{},
}

func AskJSON(tid, uid NodeID, mk MK, plr io.ReadCloser) (pl any, err error) {
	u, ok := getNode(uid)
	if !ok {
		return nil, fmt.Errorf("user node can not be found")
	}
	if mk == M_Get_Tree && u.Admin {
		tid = 1
	}
	t, ok := getNode(tid)
	if !ok {
		return nil, fmt.Errorf("target node can not be found")
	}
	if t.Owner != u && !u.Admin {
		return nil, fmt.Errorf("unauthorized message")
	}
	q := &Msg{
		Kind:    mk,
		Payload: payloadProtos[mk],
		User:    u,
	}
	if q.Payload != nil {
		err = json.NewDecoder(plr).Decode(&q.Payload)
		if err != nil {
			return nil, err
		}
	}
	a := t.AskMsg(q)
	return a.Payload, nil
}

func GetServerParms() (rp RootParms, err error) {
	a := Tree.Sys.Root.Ask(SystemUser, M_Get_Parms)
	if a.Kind == M_Error {
		return rp, a.Payload.(error)
	}
	return a.Payload.(RootParms), nil
}

func CreateUser(nm, email, pwd string) error {
	a := Tree.Sys.Users.Ask(SystemUser, M_Create, NK_User, nm,
		&UserParms{Email: email, Password: []byte(pwd)})
	if a.Kind == M_Error {
		return a.Payload.(error)
	}
	nu, _ := getNode(a.Payload.(NodeID))
	a = nu.Ask(nu, M_Create, NK_Group, "GUI")
	if a.Kind == M_Error {
		return a.Payload.(error)
	}
	slog.Info("SIGNUP created a new user.", "name", nm)
	return nil
}

func AuthUser(nm, pwd string) (uid NodeID, admin bool, err error) {
	a := Tree.Sys.Users.Ask(SystemUser, M_Get_Auth, nm, pwd)
	if a.Kind == M_Error {
		return 0, false, a.Payload.(error)
	}
	ut := a.Payload.(*Tag)
	return ut.ID, ut.Admin, nil
}

func RunGUIClient(uid NodeID, c *websocket.Conn) (err error) {
	u, ok := getNode(NodeID(uid))
	if !ok {
		return fmt.Errorf("user node can not be found")
	}
	guis := u.Ask(u, M_Get_Child, "GUI").Payload.(*Tag)
	done := make(DC)
	a := guis.Ask(u, M_Create, NK_GUI, "", c, done, u.Admin)
	if a.Kind == M_Error {
		return a.Payload.(error)
	}
	<-done
	return
}

func (t *nodeTree) LoadAndRun(sdb string) (err error) {
	if ok := FileExists(sdb); !ok {
		return fmt.Errorf("database '%s' doesn't exist", sdb)
	}
	if err = connectDB(sdb); err != nil {
		return
	}

	rh := &Head{}
	if err = Db.First(rh, 1).Error; err != nil {
		return
	}
	rh.path = "/Root"
	rh.Owner = SystemUser
	Tree.Sys.Root, err = rh.load()
	if err != nil {
		return
	}
	slog.Info("Created Root node.", "path", rh.path)

	// Still not very nice..
	var ok bool
	Tree.Sys.Users, ok = getNode(2)
	if !ok {
		return errors.New("users node can not be found")
	}
	Tree.Sys.System, ok = getNode(3)
	if !ok {
		return errors.New("users node can not be found")
	}

	a := Tree.Sys.System.Ask(SystemUser, M_Create, NK_TreeUpdater, "TreeUpdater")
	if a.Kind == M_Error {
		return fmt.Errorf("tree updater creation failed: %w", a.Payload.(error))
	}
	tu, _ := getNode(a.Payload.(NodeID))
	Tree.Sys.TreeUpdater = tu

	slog.Info("Node tree initialized.", "nnodes", Tree.LenNodes())
	return
}

func Stop() (err error) {
	a := Tree.Sys.Root.Ask(SystemUser, M_Stop)
	if a.Kind == M_Error {
		return errors.New(a.Payload.(string))
	}
	if err = CloseDB(); err != nil {
		slog.Error("couldn't close database.", "err", err)
	}
	return err
}
