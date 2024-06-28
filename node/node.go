package node

type Param struct{}

// type Credit struct {
// 	currency int
// 	amount   decimal.Decimal
// }

type Node interface {
	run()
	load(*Head) (Node, error)
	// Save() (error)
	Create() error
}

type Group struct {
	Head
}

type Account struct {
	Head
	Exchange string
}

// func (n *Head) load() error {
// 	return nil
// }

// func (n *Head) save() error {
// 	return nil
// }

// func (n *Head) path() string {
// 	return n.Name
// }

// func (n *Head) idByPath() int {
// 	nn := &Account{}
// 	nn.path()
// 	return 1
// }

// func (n *NodeHead) send(m *Msg) {
// 	n.Nerd.In <- m
// }

// func (n *NodeHead) wait(m *Msg) (r *Msg) {
// 	rc := make(chan *Msg)
// 	n.Nerd.In <- m
// 	return <-rc
// }

// func (n *NodeHead) run() {
// 	for m := range n.Nerd.In {
// 		slog.Debug("messge received", n.Nerd.Name, m)
// 	}
// }
