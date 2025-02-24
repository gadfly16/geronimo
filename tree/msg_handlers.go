package tree

import (
	"fmt"
	"log/slog"
	"strings"
)

// Message handler function
type HF func(Node, *Msg) *Msg

var msgHandlers []HF = []HF{
	M_Create: createHandler,
	M_Rename: renameHandler,
	M_Delete: deleteHandler,

	M_Get_Parms:   getParmsHandler,
	M_Get_Auth:    getAuthHandler,
	M_Get_Child:   getChildHandler,
	M_Get_Tree:    getTreeHandler,
	M_Get_Display: getDisplayHandler,

	M_Update_Parms: updateParmsHandler,

	M_Subscribe:   subscribeHandler,
	M_Unsubscribe: unsubscribeHandler,

	M_Refresh_Node: refreshNodeHandler,
	M_Refresh_Tree: refreshTreeHandler, // kind MK, owner *Tag, ...

	M_Stop:       stopHandler,
	M_UpdatePath: updatePathHandler,
}

func handleMsg(n Node, q *Msg) (a *Msg) {
	if q.User != n.head().Owner && !q.User.Admin {
		return NewErrorMsg(fmt.Errorf("unathorized message"))
	}
	slog.Debug("MSG received.", "node", n.head().path, "qk", q.KindName())
	hf := msgHandlers[q.Kind]
	if hf == nil {
		a = NewErrorMsg(fmt.Errorf("no handler for msg kind: %v", q.Kind))
	} else {
		a = hf(n, q)
		if a == nil {
			slog.Debug("MSG handled.", "node", n.head().path, "qk", q.KindName())
			return
		}
	}
	if a.Kind != M_Stop && q.resp != nil {
		q.AnswerMsg(a)
	}
	if a.Kind == M_Error {
		slog.Error("MSG resulted in error.", "node", n.head().path, "qk", q.KindName(), "err", a.Payload)
		return
	}
	slog.Debug("MSG handled.", "node", n.head().path, "qk", q.KindName(), "ak", a.KindName())
	return
}

func stopHandler(n Node, q *Msg) (a *Msg) {
	n.head().askChildrenMsg(q)
	return &Msg{Kind: M_Stop, Payload: n.head().ID}
}

func createHandler(n Node, q *Msg) (a *Msg) {
	cr, ok := n.(creator)
	if !ok {
		return NewErrorMsg(fmt.Errorf("create attempt by non creator node kind"))
	}
	pl := q.Payload.([]any)
	nnk, ok := pl[0].(NK)
	if !ok {
		nnk = NK(int(pl[0].(float64)))
	}
	if !cr.allowedChildren(nnk) {
		return NewErrorMsg(fmt.Errorf("%s kind can not be created", NKNames[nnk]))
	}
	nnm := pl[1].(string)
	if nnm == "" {
		nnm = ("New" + NKNames[nnk])
	}
	if _, ok := n.head().children[nnm]; ok {
		return NewErrorMsg(fmt.Errorf("node '%s' already exists", nnm))
	}
	nn := newNode(nnk)
	if nn == nil {
		return NewErrorMsg(fmt.Errorf("node kind '%s' not implemented yet", NKNames[nnk]))
	}
	nn.head().Kind = nnk
	nn.head().Name = nnm
	nn.head().ParentID = n.head().Tag.ID
	nn.head().Parent = n.head().Tag
	nn.head().Owner = n.head().Owner

	// First user must be set to admin
	if nnk == NK_User && len(n.head().children) == 0 {
		up, ok := pl[2].(*UserParms)
		if !ok {
			return NewErrorMsg(fmt.Errorf("wrong user payload"))
		}
		up.Admin = true
	}

	nnt, err := nn.create(pl[2:])
	if err != nil {
		return NewErrorMsg(err)
	}
	//Name might have been changed by create.
	nnm = nn.head().Name
	n.head().children[nnm] = nnt
	nn.head().path = n.head().path + "/" + nnm

	if Tree.Sys.TreeUpdater != nil {
		Tree.Sys.TreeUpdater.Notify(SystemUser, M_Refresh_Tree, M_Create, nnt, nnm, nn.head().Owner)
	}

	slog.Debug("NODE created.", "node", nn.head().path, "kind", nn.kindName())
	return &Msg{Kind: M_OK, Payload: nnt.ID}
}

func renameHandler(n Node, q *Msg) (a *Msg) {
	h := n.head()
	pl := q.Payload.([]any)
	nm := pl[0].(string)
	nnm := pl[1].(string)
	if nm == "" {
		onm := h.Name
		dbr := Db.Model(h).Where("id = ?", h.ID).Update("name", nnm)
		if dbr.Error != nil {
			return NewErrorMsg(fmt.Errorf("database error during rename: %w", dbr.Error))
		}
		h.path = strings.TrimSuffix(h.path, onm) + nnm
		h.askChildren(M_UpdatePath, q.User, h.path)
		h.updateGUIs()
		Tree.Sys.TreeUpdater.Notify(SystemUser, M_Refresh_Tree, M_Rename, h.Tag, h.Name, h.Owner)
		return oka
	}
	ch, ok := h.children[nm]
	if !ok {
		return NewErrorMsg(fmt.Errorf("node has no children named '%s'", nm))
	}
	if _, ok := h.children[nnm]; ok {
		return NewErrorMsg(fmt.Errorf("node already has a children named '%s'", nnm))
	}

	ch.Ask(q.User, M_Rename, "", nnm)

	h.children[nnm] = ch
	delete(h.children, nm)
	return oka
}

// TODO: call the node's delete method for database deletion.
func deleteHandler(n Node, q *Msg) (a *Msg) {
	nm := q.Payload.(string)
	ch, ok := n.head().children[nm]
	if !ok {
		return NewErrorMsg(fmt.Errorf("DELETE found no children named '%s'", nm))
	}
	a = ch.Ask(q.User, M_Stop)
	if a.Kind == M_Error {
		slog.Error("HEAD couldn't delete child.", "node", n.head().path, "name", nm)
		return a
	}
	Tree.Sys.TreeUpdater.Notify(SystemUser, M_Refresh_Tree, M_Delete, ch, nm, n.head().Owner)
	delete(n.head().children, nm)
	return oka
}

func getParmsHandler(n Node, q *Msg) (a *Msg) {
	prr, ok := n.(parmer)
	if !ok {
		return oka
	}
	return &Msg{Kind: M_OK, Payload: prr.getParms()}
}

func updateParmsHandler(n Node, q *Msg) (a *Msg) {
	prr, ok := n.(parmer)
	if !ok {
		return NewErrorMsg(fmt.Errorf("%s node has no parms for update", n.head().path))
	}
	pl, ok := q.Payload.(H)
	if !ok {
		return NewErrorMsg(fmt.Errorf("invalid payload type for update"))
	}
	err := prr.updateParms(pl)
	if err != nil {
		return NewErrorMsg(fmt.Errorf("failed to update parms: %w", err))
	}
	return oka
}

func getAuthHandler(n Node, q *Msg) (a *Msg) {
	usn, ok := n.(*UsersNode)
	if !ok {
		return NewErrorMsg(fmt.Errorf("only a users node can auth a user"))
	}
	ut, err := usn.authUser(q.Payload)
	if err != nil {
		return NewErrorMsg(fmt.Errorf("user authentication failed: %w", err))
	}
	return &Msg{Kind: M_OK, Payload: ut}
}

func getChildHandler(n Node, q *Msg) (a *Msg) {
	cnm := q.Payload.(string)
	ct, ok := n.head().children[cnm]
	if !ok {
		return NewErrorMsg(fmt.Errorf("children '%s' not found", cnm))
	}
	return &Msg{Kind: M_OK, Payload: ct}
}

func getTreeHandler(n Node, q *Msg) (a *Msg) {
	tree := &TreeEntry{
		ID:   n.head().ID,
		Name: n.head().Name,
		Kind: n.head().Kind,
	}
	chm := *q
	chm.resp = make(Pipe)
	for _, ch := range n.head().children {
		ch.In <- &chm
	}
	var cherr bool
	for range len(n.head().children) {
		chr := <-chm.resp
		if chr.Kind == M_Error {
			cherr = true
		} else {
			tree.Children = append(tree.Children, chr.Payload.(*TreeEntry))
		}
	}
	if cherr {
		return NewErrorMsg(fmt.Errorf("unathorized tree request downstream"))
	}
	return &Msg{Kind: M_OK, Payload: tree}
}

func getDisplayHandler(n Node, q *Msg) (a *Msg) {
	h := n.head()
	d := H{
		"Head": H{
			"ID":         h.ID,
			"Name":       h.Name,
			"Kind":       h.Kind,
			"Path":       h.path,
			"Modified":   h.CreatedAt,
			"N_children": len(h.children),
		},
	}
	dr, ok := n.(displayer)
	if ok {
		d = dr.getDisplay(d)
	}
	return &Msg{Kind: M_OK, Payload: d}
}

func subscribeHandler(n Node, q *Msg) (a *Msg) {
	pl := q.Payload.([]any)
	if pl[0].(SK) == SK_Node_Updates {
		if n.head().guiSubs == nil {
			n.head().guiSubs = make(map[*Tag]E)
		}
		gui := pl[1].(*Tag)
		n.head().guiSubs[gui] = E{}
		slog.Debug("GUI subscribed", "node", n.head().path, "gui", gui.ID)
		return nil
	}
	pd := n.(provider)
	pd.subscribe(pl)
	return nil
}

func unsubscribeHandler(n Node, q *Msg) (a *Msg) {
	pl := q.Payload.([]any)
	if pl[0].(SK) == SK_Node_Updates {
		gt := pl[1].(*Tag)
		delete(n.head().guiSubs, gt)
		slog.Debug("GUI unsubscribed", "node", n.head().path, "gui", gt.ID)
		return nil
	}
	pd := n.(provider)
	pd.unsubscribe(pl)
	return nil
}

func refreshTreeHandler(n Node, q *Msg) (a *Msg) {
	tr, ok := n.(treeRefresher)
	if !ok {
		return NewErrorMsg(fmt.Errorf("%s can not refresh the tree", n.head().path))
	}
	tr.refreshTree(q)
	return nil
}

func updatePathHandler(n Node, q *Msg) (a *Msg) {
	h := n.head()
	h.path = q.Payload.(string) + "/" + h.Name
	h.askChildren(M_UpdatePath, q.User, h.path)
	h.updateGUIs()
	return oka
}
