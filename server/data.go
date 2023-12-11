package server

import (
	"gorm.io/gorm"
	// log "github.com/sirupsen/logrus"
)

type Account struct {
	gorm.Model
	Name          string `gorm:"unique"`
	Status        string
	PasswordHash  string
	ApiPublicKey  string
	ApiPrivateKey string

	Brokers []*Broker
	// pairs   map[string]kws.TickerUpdate
	// msg     chan accountMsg
	// orders  chan order

}

type Checkpoint struct {
	gorm.Model
	Price    float64
	BrokerID uint
}

type Broker struct {
	gorm.Model
	AccountID uint
	//	Account   *Account

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

func (acc *Account) Save(db *gorm.DB) error {
	result := db.Create(acc)
	return result.Error
}
