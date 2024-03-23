package server

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	NodeRoot = iota
	NodeUser
	NodeAccount
	NodeBroker
)

const (
	StatusActive = iota
	StatusDisabled
)

var StatusKinds = map[string]uint{
	"active":   StatusActive,
	"disabled": StatusDisabled,
}

type Node struct {
	gorm.Model
	DetailType uint
	Name       string `gorm:"UNIQUE_INDEX:unique_name_per_tnode"`
	ParentID   uint   `gorm:"UNIQUE_INDEX:unique_name_per_tnode"`

	Detail Detail `gorm:"-:all"`

	parent   *Node
	children []*Node
}

// Details are only created, so the newest detail is always the current one.
// Because of this references to tree nodes are stored in the detail struct.
type Detail interface{}

type DetailModel struct {
	ID        uint
	NodeID    uint
	CreatedAt time.Time `gorm:"index"`
	Status    uint
}

type detailLoader func(*Core, *Node) error

var detailLoaders = map[uint]detailLoader{
	NodeUser:    loadUserDetail,
	NodeAccount: loadAccountDetail,
}

func (core *Core) loadChildren(parent *Node) (err error) {
	if err = core.db.Where("parent_id = ?", parent.ID).Find(&parent.children).Error; err != nil {
		return
	}

	for _, child := range parent.children {
		core.nodes[child.ID] = child
		child.parent = parent
		if err = detailLoaders[child.DetailType](core, child); err != nil {
			return
		}
		if err = core.loadChildren(child); err != nil {
			return
		}
	}
	return
}

func loadUserDetail(core *Core, node *Node) (err error) {
	user := &User{}
	if err = core.db.Omit("password").Where("node_id = ?", node.ID).First(user).Error; err != nil {
		return
	}
	node.Detail = user
	return
}

func loadAccountDetail(core *Core, node *Node) (err error) {
	acc := &Account{}
	if err = core.db.Omit("api_public_key", "api_private_key").Where("node_id = ?", node.ID).First(acc).Error; err != nil {
		return
	}
	node.Detail = acc
	return
}

func (node *Node) treeMap() (tm gin.H) {
	tm = make(gin.H)
	tm["ID"] = node.ID
	tm["DetailType"] = node.DetailType
	tm["Name"] = node.Name
	if len(node.children) == 0 {
		return
	}
	chs := []interface{}{}
	for _, chn := range node.children {
		chs = append(chs, chn.treeMap())
	}
	tm["children"] = chs
	return
}
