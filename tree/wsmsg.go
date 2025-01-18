package tree

type WsMsgKind int

// WsMsgKind is the kind of message sent over the websocket
const (
	CredentialsWsMsgKind WsMsgKind = iota
	SubscribeWsMsgKind
	UnsubscribeWsMsgKind
	UpdateWsMsgKind
	ErrorWsMsgKind
	ClientShutdownWsMsgKind
	HeartbeatWsMsgKind
	TreeNodeRenameWsMsgKind
	TreeNodeCreateWsMsgKind
	TreeNodeDeleteWsMsgKind
)

// wsMsg is the message structure for websocket communication
type wsMsg struct {
	Kind         WsMsgKind
	OTP          string
	GUIID        NodeID
	NodeID       NodeID
	NodeKind     NK
	NodeName     string
	NodeParentID NodeID
}
