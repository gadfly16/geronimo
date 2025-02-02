package tree

import (
	"time"
)

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

type ParmModel struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	HeadID    NodeID
}

type H map[string]interface{}

func (h *Head) head() *Head {
	return h
}

func (h *Head) kindName() string {
	return NKNames[h.Kind]
}
