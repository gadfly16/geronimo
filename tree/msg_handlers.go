package tree

import (
	"fmt"
)

// import (
// 	"fmt"
// )

// Message handler function
type HF func(Node, *Msg) *Msg

// Payload prototypes
var payloadProtos []any = []any{
	M_Get_Tree:    nil,
	M_Get_Display: nil,
	M_Create:      []any{NK_Noop, "", H{}},
	M_Rename:      []any{"", ""},
	M_Delete:      []any{""},
}

var msgHandlers []HF = []HF{
	M_Create: createHandler,
	M_Stop:   stopHandler,
}

// type MT struct {
// 	hf  HF
// 	mm  []*MT
// 	api bool
// 	prt []any
// }

// var msgHandlers []*MT = []*MT{
// 	M_Ok:    nil,
// 	M_Error: nil,
// 	M_Stop:  &MT{hf: stopHandler},
// }

// // var msgHandlers = []any{
// // 	NMK_Ok:    nil,
// // 	NMK_Error: nil,
// // 	NMK_Stop:  stopHandler,
// // 	NMK_Update: MM{
// // 		SMK_Create: updateCreateHandler,
// // 	},
// // 	NMK_Get: MM{
// // 		SMK_Parms: getParmsHandler,
// // 		SMK_Auth:  getAuthHandler,
// // 	},
// // }

func (h *Head) handleMsg(n Node, q *Msg) (a *Msg) {
	if q.User != h.Owner && !q.User.Admin {
		return NewErrorMsg(fmt.Errorf("unathorized message"))
	}
	hf := msgHandlers[q.Kind]
	if hf == nil {
		return NewErrorMsg(fmt.Errorf("no handler for msg kind: %v", q.Kind))
	}
	return hf(n, q)
}

// func stopHandler(n Node, q *Msg) (a *Msg) {
// 	n.head().askChildrenMsg(q)
// 	return &Msg{Kind: M_Stop, Payload: []any{n.head().ID}}
// }

// func updateCreateHandler(h *Head, q *Msg) (a *Msg) {
// 	// pl := q.Payload.([]any)
// 	return oka
// }

// func getParmsHandler(h *Head, q *Msg) (a *Msg) {
// 	return oka
// }

// func getAuthHandler(h *Head, q *Msg) (a *Msg) {
// 	return oka
// }

func createHandler(h *Head, q *Msg) (a *Msg) {
	return oka
}
