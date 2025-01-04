package core

type Kind int

const (
	RootKind Kind = iota
	GroupKind
	UserKind
	AccountKind
	TraderKind
	TreeUpdaterKind
	UsersKind
	GUIKind
)

var Kinds = map[Kind]Node{
	RootKind:        &RootNode{},
	GroupKind:       &GroupNode{},
	UserKind:        &UserNode{},
	AccountKind:     nil,
	TraderKind:      nil,
	TreeUpdaterKind: &TreeUpdaterNode{},
	UsersKind:       &UsersNode{},
	GUIKind:         &GUINode{},
}

var kindNames = map[Kind]string{
	RootKind:        "Root",
	GroupKind:       "Group",
	UserKind:        "User",
	AccountKind:     "Account",
	TraderKind:      "Broker",
	TreeUpdaterKind: "TreeUpdater",
	UsersKind:       "Users",
	GUIKind:         "GUI",
}

func (h *Head) KindName() string {
	return kindNames[h.Kind]
}

func NewNodeKind(k Kind) Node {
	switch k {
	case RootKind:
		return &RootNode{Head: &Head{Kind: RootKind, Name: "Root"}}
	case GroupKind:
		return &GroupNode{Head: &Head{Kind: GroupKind}}
	case UserKind:
		return &UserNode{Head: &Head{Kind: UserKind}, Parms: &UserParms{}}
	case AccountKind:
		return nil
	case TraderKind:
		return nil
	case TreeUpdaterKind:
		return &TreeUpdaterNode{Head: &Head{Kind: TraderKind}}
	case UsersKind:
		return &UsersNode{Head: &Head{Kind: UsersKind}, Parms: &UsersParms{}}
	case GUIKind:
		return &GUINode{Head: &Head{Kind: GUIKind}}
	default:
		return nil
	}
}
