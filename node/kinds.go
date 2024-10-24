package node

type Kind int

const (
	RootKind Kind = iota
	GroupKind
	UserKind
	AccountKind
	BrokerKind
)

var Kinds = map[Kind]Node{
	RootKind:  &RootNode{},
	GroupKind: &GroupNode{},
	UserKind:  &UserNode{},
}

var kindNames = map[Kind]string{
	RootKind:  "Root",
	GroupKind: "Group",
	UserKind:  "User",
}

func (h *Head) KindName() string {
	return kindNames[h.Kind]
}
