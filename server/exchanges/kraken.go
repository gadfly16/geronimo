package exchanges

func init() {
	exchanges["kraken"] = &kraken{}
}

type kraken struct {
}

func (ex *kraken) Connect() {}
