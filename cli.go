package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gadfly16/geronimo/server"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
)

type connection struct {
	claims *server.Claims
	client *resty.Client
	cookie string
}

type commandHandler func(server.Settings) error

var CommandHandlers = make(map[string]commandHandler)

func getTerminalPassword(prompt string) string {
	fmt.Print(prompt)
	pw, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		log.Fatal("Couldn't get password.")
	}
	return string(pw)
}

func getTerminalString(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Couldn't get input from terminal.")
	}
	return strings.TrimSuffix(text, "\n")
}

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
	if s.UserEmail == "" && server.FileExists(s.CLICookiePath) {
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
		if s.UserEmail != "" {
			user.Email = s.UserEmail
		} else {
			user.Email = getTerminalString("Login email: ")
		}
		if s.UserPassword != "" {
			user.Password = s.UserPassword
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
