package core

type Kind int

const (
	RootKind Kind = iota
	GroupKind
	UserKind
	AccountKind
	BrokerKind
	TreeUpdaterKind
)

var Kinds = map[Kind]Node{
	RootKind:        &RootNode{},
	GroupKind:       &GroupNode{},
	UserKind:        &UserNode{},
	AccountKind:     nil,
	BrokerKind:      nil,
	TreeUpdaterKind: &TreeUpdaterNode{},
}

var kindNames = map[Kind]string{
	RootKind:        "Root",
	GroupKind:       "Group",
	UserKind:        "User",
	AccountKind:     "Account",
	BrokerKind:      "Broker",
	TreeUpdaterKind: "TreeUpdater",
}

func (h *Head) KindName() string {
	return kindNames[h.Kind]
}
