package server

import (
	"github.com/dgrijalva/jwt-go"
	"gorm.io/gorm"
	// log "github.com/sirupsen/logrus"
)

type User struct {
	gorm.Model
	Email string `gorm:"unique"`
	Name  string
	Role  string

	Accounts []*Account
}

type UserSecret struct {
	gorm.Model
	UserID   uint
	Password string
}

type UserWithSecret struct {
	User   *User
	Secret *UserSecret
}

const (
	RoleUser = "user"
)

type Claims struct {
	jwt.StandardClaims
	Role string
}

type Account struct {
	gorm.Model
	UserID uint   `gorm:"UNIQUE_INDEX:unique_name_per_user"`
	Name   string `gorm:"UNIQUE_INDEX:unique_name_per_user"`
	Status string

	Brokers []*Broker
}

type AccountSecret struct {
	gorm.Model
	AccountID     uint
	ApiPublicKey  string
	ApiPrivateKey string
}

type AccountWithSecret struct {
	Account *Account
	Secret  *AccountSecret
}

type Checkpoint struct {
	gorm.Model
	Price    float64
	BrokerID uint
}

type Broker struct {
	gorm.Model
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
	//	lastOrd   *order

	// msg      chan brokerMsg
	// receipts chan order
}

// func (acc *Account) Save(db *gorm.DB) error {
// 	result := db.Create(acc)
// 	return result.Error
// }
