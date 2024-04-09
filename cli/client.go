package cli

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gadfly16/geronimo/server"
	"github.com/go-resty/resty/v2"
)

type actionData struct {
	cmd    string
	status string
	node   server.Node
	msg    server.Message
	acc    server.Account
	bro    server.Broker
}

type connection struct {
	claims *server.Claims
	client *resty.Client
	cookie string
}

type commandHandler func(server.Settings) error

var CommandHandlers = make(map[string]commandHandler)

func parseClaimsUnverified(cookie string) (claims *server.Claims, err error) {
	token, _, err := new(jwt.Parser).ParseUnverified(cookie, &server.Claims{})
	if err != nil {
		return
	}
	claims, ok := token.Claims.(*server.Claims)
	if !ok {
		return nil, errors.New("misformed claims field")
	}
	return
}

func connectServer(s *server.Settings) (conn *connection, err error) {
	conn = &connection{client: resty.New().SetBaseURL("http://" + s.HTTPAddr)}
	expirationTime := time.Time{}
	if userEmail == "" && server.FileExists(s.CLICookiePath) {
		savedCookie, err := os.ReadFile(s.CLICookiePath)
		if err != nil {
			return nil, err
		}
		conn.cookie = string(savedCookie)
		conn.claims, err = parseClaimsUnverified(conn.cookie)
		if err != nil {
			return nil, err
		}
		expirationTime = time.Unix(conn.claims.StandardClaims.ExpiresAt, 0)
	}
	if time.Now().After(expirationTime) {
		var user server.User
		if userEmail != "" {
			user.Email = userEmail
		} else {
			user.Email = getTerminalString("Login email: ")
		}
		if userPassword != "" {
			user.Password = userPassword
		} else {
			user.Password = getTerminalPassword("Login password: ")
		}

		resp, err := conn.client.R().
			SetBody(user).
			SetResult(&server.LoginResp{}).
			SetError(&server.APIError{}).
			Post(server.AuthLogin)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode() >= 400 {
			return nil, errors.New(resp.Error().(*server.APIError).Error)
		}

		conn.cookie = resp.Cookies()[0].Value
		conn.claims, err = parseClaimsUnverified(conn.cookie)
		if err != nil {
			return nil, err
		}
		os.WriteFile(s.CLICookiePath, []byte(conn.cookie), 0600)
	}
	conn.client.SetCookie(&http.Cookie{Name: "token", Value: conn.cookie})
	return conn, nil
}
