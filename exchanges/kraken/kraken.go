package kraken

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	exch "github.com/gadfly16/geronimo/exchanges"
)

const (
	pulseTimeout = 5 * time.Second
)

type nothing = struct{}

type private struct {
	pubkey    string
	privkey   string
	conn      *exch.Conn
	wsc       *websocket.Conn
	wsToken   string
	reconnect chan nothing
}

type public struct {
	conn *exch.Conn
}

var msgHandlers = map[int]func(*private, *exch.Msg) *exch.Msg{
	exch.MsgSubscribeBalance: handleSubscribeBalance,
}

func Connect(pubkey, privkey string) (conn *exch.Conn) {
	pr := private{
		pubkey:    pubkey,
		privkey:   privkey,
		conn:      exch.NewConnection(),
		reconnect: make(chan nothing, 1),
	}
	go pr.manage()
	pr.reconnect <- nothing{}
	return pr.conn
}

func (pr *private) manage() {
	pulse := time.NewTicker(pulseTimeout)

	for {
		select {
		case req := <-pr.conn.In:
			resp := msgHandlers[req.Kind](pr, req)
			pr.conn.Out <- resp
		case <-pulse.C:
			log.Debugf("ping time..")
			pulse.Stop()
			err := pr.ping()
			if err != nil {
				log.Debugf("ping failed: %v", err.Error())
				pr.reconnect <- nothing{}
				log.Debugf("reconnect request sent")
			} else {
				pulse.Reset(pulseTimeout)
			}
		case <-pr.reconnect:
			log.Debugf("reconnecting to Kraken websocket")
			pr.wsConnect()
			pulse.Reset(pulseTimeout)
		}
	}
}

func handleSubscribeBalance(pr *private, msg *exch.Msg) (resp *exch.Msg) {
	return &exch.Msg{Kind: exch.MsgOK}
}

func (pr *private) wsConnect() error {
	tokresp, err := pr.GetWebSocketsToken()
	if err != nil {
		return fmt.Errorf("couldn't get websocket token: %w", err)
	}
	log.Debugf("websocket token: %v", tokresp)

	c, _, err := websocket.Dial(context.Background(), "wss://ws.kraken.com/v2", nil)
	if err != nil {
		return fmt.Errorf("kraken ws connect: %w", err)
	}

	pr.wsc = c
	pr.wsToken = tokresp.Token
	return nil
}

func (pr *private) ping() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req := map[string]any{
		"method": "ping",
	}
	err = wsjson.Write(ctx, pr.wsc, req)
	if err != nil {
		return fmt.Errorf("kraken ws write: %w", err)
	}

	var resp any
	err = wsjson.Read(ctx, pr.wsc, &resp)
	if err != nil {
		return fmt.Errorf("kraken ws read: %w", err)
	}
	log.Infof("kraken ws response: %v", resp)

	return
}

// func runPublic() {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
// 	defer cancel()

// 	c, _, err := websocket.Dial(ctx, "wss://ws.kraken.com/v2", nil)
// 	if err != nil {
// 		log.Errorf("kraken ws connect: %v", err)
// 	}
// 	defer c.CloseNow()

// 	req := map[string]any{
// 		"method": "ping",
// 	}
// 	err = wsjson.Write(ctx, c, req)
// 	if err != nil {
// 		log.Errorf("kraken ws write: %v", err)
// 	}

// 	var resp any
// 	err = wsjson.Read(ctx, c, &resp)
// 	if err != nil {
// 		log.Errorf("kraken ws read: %v", err)
// 	}
// 	log.Infof("kraken ws response: %v", resp)

// 	c.Close(websocket.StatusNormalClosure, "")
// }
