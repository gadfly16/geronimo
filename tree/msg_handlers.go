package tree

// import (
// 	"fmt"
// )

// // Message handler function
// type HF func(Node, *Msg) *Msg

// // Message map
// type MM []MT

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

// func (h *Head) handleMsg(n Node, q *Msg) (a *Msg) {
// 	switch hl := msgHandlers[q.Kind].(type) {
// 	case nil:
// 		break
// 	case MM:
// 	out:
// 		for _, pl := range q.Payload {
// 			smk, ok := pl.(MK)
// 			if !ok {
// 				break
// 			}
// 			mme, ok := hl[smk]
// 			if !ok {
// 				break
// 			}
// 			switch mmet := mme.(type) {
// 			case HF:
// 				return mmet(n, q)
// 			case MM:
// 				hl = mmet
// 				continue
// 			default:
// 				break out
// 			}
// 		}
// 	case HF:
// 		return hl(n, q)
// 	}
// 	return NewErrorMsg(fmt.Errorf("no handler for msg kind: %v", q.Kind))
// }

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
