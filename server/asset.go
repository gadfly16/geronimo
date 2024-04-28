package server

type Asset struct {
	ID     int
	Symbol string
	Name   string
}

var Assets = map[string]Asset{
	"usd": {1, "USD", "USA Dollar"},
	"eur": {2, "EUR", "Euro"},
	"btc": {3, "BTC", "Bitcoin"},
	"eth": {4, "ETH", "Ethereum"},
	"ada": {5, "ADA", "Cardano"},
}
