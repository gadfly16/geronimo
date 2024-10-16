package node

type Kind int

const (
	RootKind Kind = iota
	GroupKind
	UserKind
	GUIKind
)

var Kinds = map[Kind]Node{
	RootKind:  &RootNode{},
	GroupKind: &GroupNode{},
	UserKind:  &UserNode{},
	GUIKind:   &GUINode{},
}

var kindNames = map[Kind]string{
	RootKind:  "Root",
	GroupKind: "Group",
	UserKind:  "User",
	GUIKind:   "GUI",
}

func (h *Head) KindName() string {
	return kindNames[h.Kind]
}
