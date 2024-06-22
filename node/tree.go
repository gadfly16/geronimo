package node

import (
	"github.com/gadfly16/geronimo/msg"
)

var t Tree

type Tree struct {
	nodes map[int]msg.Pipe
	root  msg.Pipe
}

func InitTree(m msg.Pipe) {
	t.nodes = make(map[int]msg.Pipe)
	t.nodes[0] = m
	t.root = m
}
