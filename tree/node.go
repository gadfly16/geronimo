package tree

import (
	"time"
)

type NK int

const (
	NK_Root NK = iota
	NK_Group
	NK_User
	NK_Account
	NK_Trader
	NK_TreeUpdater
	NK_Users
	NK_GUI

	NK_Noop
)

// TODO: We need to rename this to NNames.
var NKNames = map[NK]string{
	NK_Root:        "Root",
	NK_Group:       "Group",
	NK_User:        "User",
	NK_Account:     "Account",
	NK_Trader:      "Broker",
	NK_TreeUpdater: "TreeUpdater",
	NK_Users:       "Users",
	NK_GUI:         "GUI",
}
var nkTemplates = map[NK]Node{
	NK_Root:        &RootNode{},
	NK_Group:       &GroupNode{},
	NK_User:        &UserNode{},
	NK_Account:     nil,
	NK_Trader:      nil,
	NK_TreeUpdater: &TreeUpdaterNode{},
	NK_Users:       &UsersNode{},
	NK_GUI:         &GUINode{},
}

func newNode(k NK) Node {
	switch k {
	case NK_Root:
		return &RootNode{Head: &Head{Tag: &Tag{Kind: NK_Root, Owner: SystemUser}, Name: "Root"}}
	case NK_Group:
		return &GroupNode{Head: &Head{Tag: &Tag{Kind: NK_Group}}}
	case NK_User:
		return &UserNode{Head: &Head{Tag: &Tag{Kind: NK_User}}, Parms: &UserParms{}}
	case NK_Account:
		return nil
	case NK_Trader:
		return nil
	case NK_TreeUpdater:
		return &TreeUpdaterNode{Head: &Head{Tag: &Tag{Kind: NK_Trader}}}
	case NK_Users:
		return &UsersNode{Head: &Head{Tag: &Tag{Kind: NK_Users}}, Parms: &UsersParms{}}
	case NK_GUI:
		return &GUINode{Head: &Head{Tag: &Tag{Kind: NK_GUI}}}
	default:
		return nil
	}
}

type Node interface {
	create(pl []any) (*Tag, error)
	loadBody(*Head) (Node, error)
	run()

	// These methods are common to all Nodes, they are defined on Head which is
	// embedded to every Node's struct, therefore declarations are in this file.
	head() *Head
	kindName() string
}

type parmer interface {
	Node
	getParms() any
	updateParms(H) error
}

type creator interface {
	Node
	allowedChildren(NK) bool
}

type displayer interface {
	Node
	getDisplay(H) H
}

type provider interface {
	Node
	subscribe(pl []any)
	unsubscribe(pl []any)
}

type treeRefresher interface {
	Node
	refreshTree(q *Msg)
}

type ParmModel struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	HeadID    NodeID
}

func (h *Head) head() *Head {
	return h
}

func (h *Head) kindName() string {
	return NKNames[h.Kind]
}
