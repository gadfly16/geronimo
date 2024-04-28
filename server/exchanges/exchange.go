package exchanges

var exchanges = map[string]Exchange{}

type Exchange interface {
	Connect()
}
