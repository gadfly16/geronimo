# Geronimo Manual

This is the comprehensive documentation for Geronimo, a simple and reliable
trading application focused on the Stochastic Proportional Rebalancer algorithm.

## Introduction

Geronimo is designed to run on a Raspberry Pi connected to a home-grade router.
This provides a cost-effective solution that offers decent availability
alongside strong privacy and security guarantees for users. The system connects
to users' CEX accounts and facilitates trades according to their configured SPR
algorithm settings.

The Raspberry Pi executes the back-end code while users access the web interface
from their phone or desktop connected to the same router. With DDNS
configuration, the appliance can also be made accessible from outside the local
network.

### The SPR Algorithm

Most trading bots allow users to define, backtest, and run custom algorithms.
Geronimo takes a different approach by focusing on a single trading strategy
called the Stochastic Proportional Rebalancer. This concentrated focus enables
us to provide a unique user experience that appeals to traders seeking a balance
between ease of use, peace of mind, calculable risks and costs, and competitive
yields.

## Usage Guide

In Geronimo, everything is a node, and these nodes form a tree structure where
each node can have children. For example, an `Account` node represents a
connection to a CEX account, and users can create `Trader` nodes as its children
to facilitate autonomous trading on that account. The third type of node is
`Group`, which enables tree reorganization and creates passive fund partitions
called pockets. The final node type users need to understand is `User`, where
user-level settings are configured.

`Account` nodes _pump_ funds into the tree, while `Group` and `Trader` nodes
_drain_ funds and _pump_ them downstream. The GUI displays accumulated
downstream statistics in each node's info section, providing both detailed
information and high-level oversight of the tree's state.

Nodes can also have _parameter_—values that users can set to determine the
node's behavior.

### Experience Contracts

The system must comply with well-defined behavioral contracts to be considered
correct.

#### UI Consistency

The user interface connects only to the back-end and does not communicate with
any other parties involved in trading. This prevents the front-end from seeing a
different reality than the back-end. In other words, users can always be
confident that what they see on the interface reflects how the back-end sees the
world.

#### Actuality

The front-end continuously monitors the health of its connection to the server
and reports this status in the bottom status bar. When the connection is
healthy, the status light is green; when broken, it's red. When the connection
is healthy, users can be confident that the displayed parameter and info values
are consistent with the back-end state. The system complies with this contract
by continuously updating all displayed topological, info, and parameter values,
and by reloading the complete display state after every reconnection.

#### User Changes

Topological and parameter changes initiated by users are only updated in the UI
after the backend reports successful commitment to persistent storage and
resends the parameter as an update. When the UI displays the updated value,
users can be confident that the update will persist after a restart.

#### Parameter History

The system stores parameter values as timestamped lists and displays the latest
state by default. This means the UI can revisit earlier node states to some
extent. Unfortunately, v1.0 will most likely not support revisiting prior
topological states, which is semi-intentional. While this inherently makes
revisiting previous states a dangerous endeavor, knowing older node states is
still considered essential. This design has the side effect that when a user
deletes a node, that node is permanently removed along with all related
information from persistent storage, including all parameter history.

## Build Instructions

### Dependencies

You need a working Go environment and the following dependencies to build the
back-end executable and the distributable GUI:

```
sudo pacman -Syyu protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
sudo npm install -g ts-proto
sudo pacman -Syyu esbuild
```

### Build and Run

Peek into `generate.go` in the root of the project to understand the build
process. The following line run from the project root should generate, build and
execute the back-end:

```
go generate . && go build . && ./geronimo serve
```
