package core

type NodeID int

type Tag struct {
	ID       NodeID
	Kind     Kind
	Name     string
	Node     Pipe
	ParentID NodeID
	OwnerID  NodeID
}
