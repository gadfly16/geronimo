package server

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Account struct {
	DetailModel
	Exchange      string
	APIPublicKey  string
	APIPrivateKey string
}

func (account *Account) displayData() (display gin.H) {
	display = gin.H{}
	display["Detail"] = account
	return
}

func (account *Account) run() error {
	log.Debug("Running account: ", core.nodes[account.NodeID].Name)
	return nil
}

// func (core *Core) saveAccount(aws *AccountWithSecret) error {
// 	tx := core.db.Create(aws.Account)
// 	if tx.Error != nil {
// 		return tx.Error
// 	}
// 	tx =
// 	return result.Error
// }

// "gorm.io/gorm"
// log "github.com/sirupsen/logrus"

// type accountMsg struct{}

// func (acc *account) init() {
// 	log.Info("Initializing account: ", acc.name)

// 	acc.db = openDB()

// 	// Load and initialize brokers
// 	sqlStmt := `
// 		SELECT
// 			id, name, pair, status, minWait, maxWait,
// 			highLimit, lowLimit, delta, offset, base, quote, fee
// 		FROM broker
// 		WHERE accountId = $1
// 		AND status = "active"`
// 	rows, err := acc.db.Query(sqlStmt, acc.id)
// 	if err != nil {
// 		log.Fatal("Couldn't get brokers for account: ", err)
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var bro broker
// 		if err := rows.Scan(&bro.id, &bro.name, &bro.pair,
// 			&bro.status, &bro.minWait, &bro.maxWait, &bro.highLimit, &bro.lowLimit,
// 			&bro.delta, &bro.offset, &bro.base, &bro.quote, &bro.fee); err != nil {
// 			log.Fatal("Couldn't scan broker for running: ", err)
// 		}
// 		bro.acc = acc
// 		bro.msg = make(chan brokerMsg)
// 		bro.receipts = make(chan order)

// 		bro.init()

// 		acc.brokers[bro.id] = &bro
// 		acc.pairs[bro.pair] = kws.TickerUpdate{}
// 	}
// 	if err = rows.Err(); err != nil {
// 		log.Fatal(err)
// 	}

// }

// func getAccountID(db *sql.DB, name string) int64 {
// 	var accountId int64
// 	sqlStmt := "SELECT id FROM account WHERE name = $1"
// 	if err := db.QueryRow(sqlStmt, name).Scan(&accountId); err != nil {
// 		log.Fatal("Couldn't get account ID:", err)
// 	}
// 	log.Debug("Account found: ", accountId)
// 	return accountId
// }

// func (acc *account) decryptKeys() {
// 	password := getTerminalString(fmt.Sprintf("Enter Password for account `%s`: ", acc.name))
// 	if acc.pwhash != hashPassword(password) {
// 		log.Fatal("Wrong password.")
// 	}
// 	acc.apiPublicKey = decryptString(password, acc.name, acc.apiPublicKey)
// 	acc.apiPrivateKey = decryptString(password, acc.name, acc.apiPrivateKey)
// 	log.Debug("Password checked.")
// }

// // Runs the goroutine managing the account
// func (acc *account) run() {
// 	log.Info("Running account: ", acc.name)

// 	krakenPub := krakenConnectPublic()
// 	krakenPriv := krakenConnectPrivate(acc)

// 	db := openDB()
// 	defer db.Close()

// 	// Get active brokers

// 	krakenSubscribeTickers(krakenPub, acc.pairs)
// 	krakenSubscribeOpenOrders(krakenPriv)
// 	krakenSubscribeOwnTrades(krakenPriv)

// 	// Initialization done, registering account in global map
// 	acc.msg = make(chan accountMsg)
// 	acc.orders = make(chan order)

// 	brokersStarted := false
// 	tradesBooked := false
// 	pubUpdates := krakenPub.Listen()
// 	privUpdates := krakenPriv.Listen()
// 	for {
// 		select {
// 		case update := <-pubUpdates:
// 			switch data := update.Data.(type) {
// 			case kws.TickerUpdate:
// 				acc.pairs[update.Pair] = data
// 				log.Debugf("Updated ticker: %v (%v)", update.Pair, data.Ask.Price)
// 				if !brokersStarted && tradesBooked {
// 					log.Debugf("First ticker arrived and trades booked, starting brokers for account: `%v`.", acc.name)
// 					for _, bro := range acc.brokers {
// 						go bro.run()
// 					}
// 					brokersStarted = true
// 				}
// 			}
// 		case update := <-privUpdates:
// 			switch data := update.Data.(type) {
// 			case kws.OpenOrdersUpdate:
// 				for _, openOrdDict := range data {
// 					for id, openOrd := range openOrdDict {
// 						log.Debugf("Open order update: %v: %+v", id, openOrd)
// 						updateOrder(db, openOrd, id)
// 					}
// 				}
// 			case kws.OwnTradesUpdate:
// 				for _, oTrdDict := range data {
// 					for id, ownTrd := range oTrdDict {
// 						if tradeExists(db, id) {
// 							log.Debugf("Trade already booked: %v", id)
// 							continue
// 						}
// 						trdOrd := getOrderById(db, acc, ownTrd.OrderID)
// 						if trdOrd == nil {
// 							log.Debugf("Trade doesn't belong to any order: %v", id)
// 							continue
// 						}
// 						log.Debugf("Bookkeeping trade: %v", ownTrd)
// 						acc.brokers[trdOrd.bro.id].bookTrade(db, id, ownTrd, trdOrd)
// 					}
// 				}
// 				tradesBooked = true
// 			}
// 		case ord := <-acc.orders:
// 			if ord.midPrice == 0 {
// 				// Broker asking for current midPrice
// 				ticker := acc.pairs[ord.bro.pair]
// 				askPrice, err := ticker.Ask.Price.Float64()
// 				if err != nil {
// 					log.Fatal("Couldn't convert ask price.")
// 				}
// 				bidPrice, err := ticker.Bid.Price.Float64()
// 				if err != nil {
// 					log.Fatal("Couldn't convert bid price.")
// 				}
// 				ord.midPrice = (askPrice + bidPrice) / 2
// 				ord.bro.receipts <- ord
// 			} else {
// 				log.Debug("Received priced order from broker: ", ord.bro.name)
// 				saveOrder(db, &ord)
// 				openOrderIds := getOpenOrderIds(db, ord.bro.id)
// 				if len(openOrderIds) != 0 {
// 					err := krakenPriv.CancelOrder(openOrderIds)
// 					if err != nil {
// 						log.Fatalf("Couldn't cancel open orders: %v", openOrderIds)
// 					}
// 				}
// 				log.Infof("Canceled orders: %v", openOrderIds)
// 				ordType := "buy"
// 				if ord.volume < 0 {
// 					ordType = "sell"
// 				}
// 				req := kws.AddOrderRequest{
// 					Ordertype: "limit",
// 					Pair:      ord.bro.pair,
// 					Price:     fmt.Sprintf("%f", ord.price),
// 					Type:      ordType,
// 					Volume:    fmt.Sprintf("%f", math.Abs(ord.volume)),
// 					UserRef:   fmt.Sprintf("%d", ord.userRef),
// 				}
// 				err := krakenPriv.AddOrder(req)
// 				if err != nil {
// 					log.Fatal("Couldn't place order: ", ord)
// 				}
// 				log.Infof("Placed order: %+v", req)
// 			}
// 		}
// 	}
// }
