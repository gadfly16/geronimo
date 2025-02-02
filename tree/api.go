package tree

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/coder/websocket"
)

// Payload prototypes
var payloadProtos []any = []any{
	M_Get_Tree:    nil,
	M_Get_Display: nil,
	M_Create:      []any{NK_Noop, "", H{}},
	M_Rename:      []any{"", ""},
	M_Delete:      []any{""},
}

func AskJSON(tid, uid NodeID, mk MK, plr io.ReadCloser) (pl any, err error) {
	t, ok := getNode(tid)
	if !ok {
		return nil, fmt.Errorf("target node can not be found")
	}
	u, ok := getNode(uid)
	if !ok {
		return nil, fmt.Errorf("user node can not be found")
	}
	if t.Owner != u && !u.Admin {
		return nil, fmt.Errorf("unauthorized message")
	}
	q := &Msg{
		Kind:    mk,
		Payload: payloadProtos[mk],
	}
	if q.Payload != nil {
		err = json.NewDecoder(plr).Decode(q.Payload)
		if err != nil {
			return nil, err
		}
	}
	// If the request is to get the tree, tree is served from the root node
	if q.Kind == M_Get_Tree && u.Admin {
		tid = 1
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
	nu := a.Payload.(*Tag)
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
