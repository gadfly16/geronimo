package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/gadfly16/geronimo/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	act        actionData
	runtimeErr bool
)

type actionData struct {
	cmd    string
	status string
	node   server.Node
	msg    server.Message
	acc    server.Account
	bro    server.Broker
	group  server.Group
	pocket server.Pocket
}

func cliError(errMsg string) {
	log.Error(errMsg)
	runtimeErr = true
}

func addActionFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&act.msg.Path, "path", "p", "", "path of the action")
}

func addObjectFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&act.status, "status", "s", "active", "status of the object")
}

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
