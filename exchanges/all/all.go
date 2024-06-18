package all

import (
	exch "github.com/gadfly16/geronimo/exchanges"

	"github.com/gadfly16/geronimo/exchanges/kraken"
)

var Connect = map[string]func(string, string) *exch.Conn{
	"kraken": kraken.Connect,
}
