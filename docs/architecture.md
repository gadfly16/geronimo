# Geronimo's Architecture

## Introduction

Geronimo impements a reliable and persistent _tree_ of concurrently running
_nodes_, that communicate with each other through message passing.

### Performance

Nerd is fast enough to build useful applications on a RPI3B+.

### Reliable

Nerd provides strong data consistency guarantees. Changes to Node _parameters_
are atomic and due to the way the tree works, all queries return a consistent
view of the tree state.

### Persistent

When a user experiences a change that also means that that change is saved to
permanent storage and will be present after a server failure.

## Architecture

### The Tree

The basic building blocks of Nerd are _nodes_, _messages_ and _satelites_. Nodes
are independent goroutines forming a tree and they're only communicating through
message passing. Every node has an `in` channel that receives messages from
other nodes. Because parent nodes can query their siblings recursively it is
forbidden for a node to send messages to the in channel of any upstream nodes.
