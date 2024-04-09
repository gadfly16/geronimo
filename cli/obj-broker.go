package cli

import (
	"errors"

	"github.com/gadfly16/geronimo/server"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

func init() {
	addObjectFlags(brokerCmd)
	// brokerCmd.Flags().StringVar(&act.broker.Name, "n", "defaultBroker", "Name of the new broker.")
	// brokerCmd.Flags().StringVar(&accountName, "a", "defaultAccount", "Name of the account the new broker belongs to.")
	brokerCmd.Flags().StringVarP(&act.bro.Pair, "market", "m", "ADA/EUR", "the market's name the broker trades on")
	brokerCmd.Flags().Float64VarP(&act.bro.Base, "base-amount", "b", 0, "amount of 'base' currency handled by the broker")
	brokerCmd.Flags().Float64VarP(&act.bro.Quote, "quote-amount", "q", 0, "amount of 'quote' currency handled by the broker")
	brokerCmd.Flags().Float64VarP(&act.bro.MinWait, "min-wait", "w", 3600, "minimum wait time between checks in seconds")
	brokerCmd.Flags().Float64VarP(&act.bro.MaxWait, "max-wait", "x", 10800, "maximum wait time between checks in seconds")
	brokerCmd.Flags().Float64VarP(&act.bro.HighLimit, "top-limit", "t", 5, "high limit")
	brokerCmd.Flags().Float64VarP(&act.bro.LowLimit, "low-limit", "l", 0.2, "low limit")
	brokerCmd.Flags().Float64VarP(&act.bro.Delta, "delta", "d", 0.04, "minimum price delta between trades")
	brokerCmd.Flags().Float64VarP(&act.bro.Offset, "offset", "o", 0.0025, "offset of limit trades from current price")
	createCmd.AddCommand(brokerCmd)
}

var brokerCmd = &cobra.Command{
	Use:   "broker",
	Short: "defines an broker to operate on",
	Long: `The 'broker' command defines the details of a broker object
			so that different actions can be performed on them.`,
	Run: runBroker,
}

func runBroker(cmd *cobra.Command, args []string) {
	log.Debugln("Executing", act.cmd, "on broker object.")
	act.bro.Status = server.StatusKinds[act.status]

	act.node.DetailType = server.NodeBroker
	act.node.Detail = &act.bro
	act.msg.Payload = &act.node

	route := "/api" + server.APICreate + "/broker"
	conn, err := connectServer(&s)
	if err != nil {
		cliError(err)
		return
	}

	log.Debugln(act.msg, act.msg.Payload, act.msg.Payload.(*server.Node).Detail)

	resp, err := conn.client.R().
		SetBody(&act.msg).
		SetError(&server.APIError{}).
		Post(route)
	if err != nil {
		cliError(err)
		return
	}
	if resp.StatusCode() >= 400 {
		cliError(errors.New(resp.Error().(*server.APIError).Error))
		return
	}
}
