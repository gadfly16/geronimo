package server

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	NodeRoot = iota
	NodeUser
	NodeAccount
	NodeBroker
	NodeGroup
	NodePocket
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

	Detail Detailer `gorm:"-:all"`

	parent   *Node
	children map[string]*Node
}

type Detailer interface {
	displayData() gin.H
	run() error
}

// Details are inmutable, the newest detail is always the current one.
// Because of this references to tree nodes are stored in the detail struct.
type DetailModel struct {
	ID        uint
	NodeID    uint
	CreatedAt time.Time `gorm:"index"`
	Status    uint
}

type detailLoader func(*Node) error

var detailLoaders = map[uint]detailLoader{
	NodeUser:    loadUserDetail,
	NodeAccount: loadAccountDetail,
	NodeBroker:  loadBrokerDetail,
}

func (node *Node) path() string {
	path := "/" + node.Name
	parent := node.parent
	for parent.DetailType != NodeRoot {
		path = "/" + parent.Name + path
		parent = parent.parent
	}
	return path
}

func (core *Core) getNode(id uint, user *Node) (*Node, error) {
	userDetail := user.Detail.(*User)
	node, ok := core.nodes[id]
	if !ok {
		return nil, errors.New("node doesn't exists")
	}
	if userDetail.Role != RoleAdmin && node.findUpstreamClass(NodeUser) != user {
		return nil, errors.New("node doesn't belong to user")
	}
	return node, nil
}

func (node *Node) findUpstreamClass(class uint) *Node {
	if node.parent == nil {
		return nil
	}
	if node.DetailType == class {
		return node
	}
	return node.parent.findUpstreamClass(class)
}

func (node *Node) run() (err error) {
	err = node.Detail.run()
	if err != nil {
		return
	}
	for _, ch := range node.children {
		err = ch.run()
		if err != nil {
			return
		}
	}
	return
}

func (node *Node) loadChildren() (err error) {
	node.children = make(map[string]*Node)
	chs := []*Node{}
	if err = core.db.Where("parent_id = ?", node.ID).Find(&chs).Error; err != nil {
		return
	}

	for _, child := range chs {
		node.children[child.Name] = child
		core.nodes[child.ID] = child
		child.parent = node
		if err = detailLoaders[child.DetailType](child); err != nil {
			return
		}
		if err = child.loadChildren(); err != nil {
			return
		}
	}
	return
}

func loadUserDetail(node *Node) (err error) {
	user := &User{}
	if err = core.db.Omit("password").Where("node_id = ?", node.ID).Order("created_at desc").Take(user).Error; err != nil {
		return
	}
	node.Detail = user
	return
}

func loadAccountDetail(node *Node) (err error) {
	acc := &Account{}
	if err = core.db.Omit("api_public_key", "api_private_key").Where("node_id = ?", node.ID).Order("created_at desc").Take(acc).Error; err != nil {
		return
	}
	node.Detail = acc
	return
}

func loadBrokerDetail(node *Node) (err error) {
	bro := &Broker{}
	if err = core.db.Where("node_id = ?", node.ID).Order("created_at desc").Take(bro).Error; err != nil {
		return
	}
	node.Detail = bro
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

func (node *Node) display() (display gin.H) {
	display = gin.H{}
	display["Name"] = node.Name
	display["DetailType"] = node.DetailType
	display["ID"] = node.ID
	display["Path"] = node.path()
	for k, v := range node.Detail.displayData() {
		display[k] = v
	}
	return
}

func find(parent *Node, path string, user *Node) (node *Node) {
	return findNode(parent, strings.Split(path[1:], "/"), user)
}

func findParent(parent *Node, path string, user *Node) (node *Node) {
	pathSlice := strings.Split(path[1:], "/")
	return findNode(parent, pathSlice[:len(pathSlice)-1], user)
}

func findNode(parent *Node, path []string, user *Node) (node *Node) {
	userDetail := user.Detail.(*User)
	node, ok := parent.children[path[0]]
	if !ok {
		return
	}
	if node.DetailType == NodeUser {
		if userDetail.Role != "admin" && node.ID != userDetail.NodeID {
			return nil
		}
	}
	if len(path) > 1 {
		return findNode(node, path[1:], user)
	}
	return
}

// Returns the name of the node from its path
func name(path string) string {
	pathSlice := strings.Split(path[1:], "/")
	return pathSlice[len(pathSlice)-1]
}
