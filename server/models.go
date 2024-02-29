package server

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/gorm"
	// log "github.com/sirupsen/logrus"
)

const (
	NodeRoot = iota
	NodeUser
	NodeAccount
	NodeBroker
)

type Detail interface{}

type treeNode struct {
	gorm.Model
	DetailType uint
	ParentID   uint
	Parent     *treeNode
	Children   []*treeNode

	detail *Detail
}

type detailModel struct {
	ID         uint      `gorm:"primarykey"`
	CreatedAt  time.Time `gorm:"index"`
	TreeNodeID uint
	Name       string
	TreeNode   *treeNode
}

type UserDetail struct {
	detailModel
	Email string `gorm:"unique"`
	Name  string
	Role  string
}

type UserSecret struct {
	UserID   uint `gorm:"primarykey"`
	Password string
}

type UserWithSecret struct {
	User   *UserDetail
	Secret *UserSecret
}

const (
	RoleUser = "user"
)

type Claims struct {
	jwt.StandardClaims
	Role string
}

type AccountDetail struct {
	detailModel
	UserID uint   `gorm:"UNIQUE_INDEX:unique_name_per_user"`
	Name   string `gorm:"UNIQUE_INDEX:unique_name_per_user"`
	Status string
}

type AccountSecret struct {
	AccountID     uint
	ApiPublicKey  string
	ApiPrivateKey string
}

type AccountWithSecret struct {
	Account *AccountDetail
	Secret  *AccountSecret
}

type Checkpoint struct {
	CreatedAt time.Time
	BrokerID  uint
	Price     float64
}

type BrokerDetail struct {
	detailModel
	AccountID uint

	Name      string
	Pair      string
	Status    string
	MinWait   float64
	MaxWait   float64
	HighLimit float64
	LowLimit  float64
	Delta     float64
	Offset    float64
	Base      float64
	Quote     float64
	Fee       float64

	LastCheck *Checkpoint
}
