package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"strconv"

	"github.com/gadfly16/geronimo/server"
)

func init() {
	CommandHandlers["get-state"] = getStateCLI
}

func getStateCLI(s server.Settings) error {
	var userID uint
	flags := flag.NewFlagSet("show", flag.ExitOnError)
	flags.UintVar(&userID, "user-id", 0, "User ID of the owner of the account.")
	flags.Parse(flag.Args()[1:])

	conn, err := connectServer(&s)
	if err != nil {
		return err
	}

	if userID == 0 {
		uid, err := strconv.Atoi(conn.claims.StandardClaims.Subject)
		if err != nil {
			return err
		}
		userID = uint(uid)
	}

	resp, err := conn.client.R().
		SetError(&server.APIError{}).
		SetQueryParam("userid", strconv.Itoa(int(userID))).
		Get("/api" + server.APIState)
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
