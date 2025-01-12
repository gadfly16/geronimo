package core

type WsMsgKind int

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

type wsMsg struct {
	Kind         WsMsgKind
	OTP          string
	GUIID        NodeID
	NodeID       NodeID
	NodeKind     Kind
	NodeName     string
	NodeParentID NodeID
}
