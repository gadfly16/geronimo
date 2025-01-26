# The Nerd Architecture

## Introduction
Nerd impements a performant, reliable and persistent *tree* of concurrently running *nodes*, that communicate with eachother through message passing.

### Performance
Nerd is fast enough to build useful applications on a RPI3B+.

### Reliable
Nerd provides strong data consistency guarantees. Changes to Node *attributes* are atomic and due to the way the tree works all queries return a consistent view of the tree state.

### Persistent
When a user experiences a change that also means that that change is saved to permanent storage and will be present after a server failure.
