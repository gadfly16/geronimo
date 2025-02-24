package tree

type WSMK int

// WsMsgKind is the kind of message sent over the websocket
const (
	WSM_Credentials WSMK = iota
	WSM_Subscribe
	WSM_Unsubscribe
	WSM_Update
	WSM_Error
	WSM_ClientShutdown
	WSM_Heartbeat
	WSM_Renamed
	WSM_Created
	WSM_Deleted
)

// wsMsg is the message structure for websocket communication
type wsMsg struct {
	Kind         WSMK
	OTP          string
	GUIID        NodeID
	NodeID       NodeID
	NodeKind     NK
	NodeName     string
	NodeParentID NodeID
}
