package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"

	kws "github.com/aopoltorzhicky/go_kraken/websocket"
	log "github.com/sirupsen/logrus"
)

type account struct {
	id            int64
	name          string
	password      string
	apiPublicKey  string
	apiPrivateKey string
}

func getAccountID(db *sql.DB, name string) int64 {
	var accountId int64
	sqlStmt := "SELECT id FROM account WHERE name = $1"
	if err := db.QueryRow(sqlStmt, name).Scan(&accountId); err != nil {
		log.Fatal("Couldn't get account ID:", err)
	}
	log.Debug("Account found: ", accountId)
	return accountId
}

func setAccountKey(acc *account) {
	fmt.Printf("Enter Password for account `%s`: ", acc.name)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("Couldn't get password.")
	}

	log.Debug("Pw: ", string(bytePassword))
}

func runBookkeeper(acc account) {
	log.Debug("Running bookkeeper for: ", acc.name)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	kraken := kws.NewKraken(kws.ProdBaseURL)
	if err := kraken.Connect(); err != nil {
		log.Fatalf("Error connecting to web socket: %s", err.Error())
	}
	// if err := kraken.Authenticate(os.Getenv("KRAKEN_API_KEY"), os.Getenv("KRAKEN_SECRET")); err != nil {
	// 	log.Fatalf("Authenticate error: %s", err.Error())
	// }

	db := openDB()
	defer db.Close()

	sqlStmt := `
		SELECT * FROM broker
		WHERE accountId = $1
		AND status = "active"`
	rows, err := db.Query(sqlStmt, acc.id)
	if err != nil {
		log.Fatal("Couldn't get brokers for account: ", err)
	}
	defer rows.Close()

	bros := map[int64]broker{}
	tickers := map[string]kws.TickerUpdate{}
	tickerList := []string{}

	for rows.Next() {
		var bro broker
		if err := rows.Scan(&bro.id, &bro.accountId, &bro.name, &bro.pair,
			&bro.status, &bro.minWait, &bro.maxWait, &bro.highLimit, &bro.lowLimit,
			&bro.delta, &bro.offset, &bro.base, &bro.quote); err != nil {
			log.Fatal("Couldn't scan broker for running: ", err)
		}
		bros[bro.id] = bro
		tickers[bro.pair] = kws.TickerUpdate{}
		tickerList = append(tickerList, bro.pair)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	if err := kraken.SubscribeTicker(tickerList); err != nil {
		log.Fatalf("SubscribeTicker error: %s", err.Error())
	}

	orders := make(chan order)
	receipts := map[int64](chan order){}

	brokersStarted := false
	updates := kraken.Listen()
	for {
		select {
		case <-signals:
			log.Warn("Stopping...")
			if err := kraken.Close(); err != nil {
				log.Fatal(err)
			}
			return
		case update := <-updates:
			switch data := update.Data.(type) {
			case kws.TickerUpdate:
				tickers[update.Pair] = data
				log.Debugf("Updated ticker: %v (%v)", update.Pair, data.Ask.Price)
				if !brokersStarted {
					log.Debugf("First ticker arrived, brokers can be started for `%s`.", acc.name)
					for _, bro := range bros {
						receipt := make(chan order)
						receipts[bro.id] = receipt
						go runBroker(bro, orders, receipt)
						brokersStarted = true
					}
				}
			}
		case ord := <-orders:
			if ord.midPrice == 0 {
				// Broker asking for current midPrice
				ticker := tickers[bros[ord.brokerId].pair]
				askPrice, err := ticker.Ask.Price.Float64()
				if err != nil {
					log.Fatal("Couldn't convert ask price.")
				}
				bidPrice, err := ticker.Bid.Price.Float64()
				if err != nil {
					log.Fatal("Couldn't convert bid price.")
				}
				ord.midPrice = (askPrice + bidPrice) / 2
				receipts[ord.brokerId] <- ord
			}
		}
	}
}
