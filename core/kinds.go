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

var KindTemplates = map[Kind]Node{
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
		return &RootNode{Head: &Head{Tag: &Tag{Kind: RootKind}, Name: "Root", Owner: SystemUser}}
	case GroupKind:
		return &GroupNode{Head: &Head{Tag: &Tag{Kind: GroupKind}}}
	case UserKind:
		return &UserNode{Head: &Head{Tag: &Tag{Kind: UserKind}}, Parms: &UserParms{}}
	case AccountKind:
		return nil
	case TraderKind:
		return nil
	case TreeUpdaterKind:
		return &TreeUpdaterNode{Head: &Head{Tag: &Tag{Kind: TraderKind}}}
	case UsersKind:
		return &UsersNode{Head: &Head{Tag: &Tag{Kind: UsersKind}}, Parms: &UsersParms{}}
	case GUIKind:
		return &GUINode{Head: &Head{Tag: &Tag{Kind: GUIKind}}}
	default:
		return nil
	}
}
