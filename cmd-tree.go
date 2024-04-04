package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/gadfly16/geronimo/server"
)

func init() {
	parser.AddCommand(
		"tree",
		"gets the object tree from the server",
		"The tree command gets the tree structure from the server.",
		&treeOptions{})
}

type treeOptions struct {
	UserID uint `short:"u"`
}

func (treeOpts *treeOptions) Execute(args []string) error {
	s.Init()
	conn, err := connectServer(&s)
	if err != nil {
		return err
	}

	if treeOpts.UserID == 0 {
		uid, err := strconv.Atoi(conn.claims.StandardClaims.Subject)
		if err != nil {
			return err
		}
		treeOpts.UserID = uint(uid)
	}

	resp, err := conn.client.R().
		SetError(&server.APIError{}).
		SetQueryParam("userid", strconv.Itoa(int(treeOpts.UserID))).
		Get("/api" + server.APITree)
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 400 {
		return errors.New(resp.Error().(*server.APIError).Error)
	}

	state := map[string]any{}
	if err = json.Unmarshal(resp.Body(), &state); err != nil {
		return err
	}
	output, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
