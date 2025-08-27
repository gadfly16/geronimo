# Geronimo Manual

This is the comprehensive documentation for Geronimo, a simple and reliable
trading application focused on the Stochastic Rational Rebalancer (SRR)
algorithm.

## Introduction

Geronimo is designed to run on a _household internet agent_ (HIA) connected to
the household's router.

HIA is a new device category we introduced at `senki.se` to refer to small, low
consumption headless computers connected to household routers. You can think of
them as the server equivalents of mobile phones. Hardware wise a Raspberry Pi 3
can provide several useful HIA services, and it can do it under very different
privacy and security conditions than other solutions like VPSs or cloud
solutions.

A Geronimo instance installed on a HIA can connect to the users' CEX accounts
and facilitate trades autonomously according to its configuration.

The HIA executes the back-end code while users access the web interface from
their phone or desktop connected to the same router. With DDNS configuration,
the web interface can also be made accessible from outside the local network.

## The SRR Algorithm

Most trading bots allow users to define, backtest, and run custom algorithms.
Geronimo takes a different approach by focusing on a single trading strategy
called the Stochastic Rational Rebalancer. This concentrated focus enables us to
provide a unique user experience that appeals to traders seeking a balance
between ease of use, peace of mind, calculable risks and costs, and competitive
yields.

The algorithm is strikingly simple yet it provides very nice characteristics.
This includes competitive yield calculably proportional to the volatility on the
underlying market and straightforward risk analysis through restraining from any
prediction attempts.

Due to its simplicity, SRR is easily extensible by automatically tuning its core
parameters over time. We don't discuss these adaptation possibilities here,
instead we are focusing on the basic concept and its analysis.

### Basic Principles

The SRR trades on a single market and must know how much it holds from each side
of the pair.

- `B`: base amount
- `Q`: quote amount

The users only predict the range they expect the price to stay in the
foreseeable future.

- `L`: low price limit
- `H`: high price limit

Knowing the high and low limits and the current price the desirable _ratio_ can
be calculated for any given price point.

- `P`: current price
- `R = (ln(P) - ln(L)) / (ln(H) - ln(L))`: ratio

This formula gives back `0` when `P` is equal or lower than `L`, `1` when `P` is
equal or higher than `H` and `0.5` at the geometric mean of `L` and `H`. A ratio
of `0` means that _all_ funds under the provision of the trader should be
converted to the base currency, while a ratio of `1` means the opposite, that
all funds should be converted to the quote. If the price is in between `L` and
`H` the proportion would be a number between `0` and `1` that gradually changes
between the two extremes.

Translating this to trader's terms we could say: I would like to convert _all_
of my BTC to USD when the price reaches the high limit for example $1M, but I
want to convert all of my USD to BTC when the price reaches the low limit for
example $10000. If the price is at the geometric mean - in this example
$100000 - I want the value of my two pairs to be equal. And I want to
continuously rebalance my funds to reflect the ratio the current price dictates.

To make the algorithm profitable a threshold must be introduced, the rule being
that a rebalancing attempt is only made when the ratio of the necessary trade's
value and the total value of all funds - the so called imbalance - is above a
user specified value. To simplify the imbalance formula we introduced the total
value.

- `T`: threshold
- `V = B * P + Q`: total value of funds
- `I = |R * V - Q| / V`: imbalance

### SRR Toy

This little python snippet sets up a simple SRR environment that you can play
with in an interactive shell:

```python
import math
ln = math.log

# Necessary data
B = 100       # Base funds
Q = 100       # Quote funds
P = 1         # Current price

# Parameters
L = .1        # Low limit
H = 10        # High limit
T = 0.05      # Imbalance threshold

# Formulas
def V():
    return B * P + Q

def R():
    return (ln(P)-ln(L)) / (ln(H) - ln(L))

def I():
    return abs(V() * R() - Q) / V()

def status():
    print("B:{:.3F} Q:{:.3F} P:{} L:{} H:{} T:{} V:{:.3F} R:{:.3f} I:{:.3f} ".format(B,Q,P,L,H,T,V(),R(),I()))

def rebalance():
    global Q, B
    v = V()
    r = R()
    Q = v * r
    B = (v - Q) / P
    status()
```

### Timing

The last question left is _when_ do we check the price and decide if we want to
initiate a rebalance or not. The basic answer to this question is that to be
profitable _checks_ should happen regularly. That's also easy to prove that
either increasing or decreasing the frequency of checks to the extremes damages
yields and in case of high frequencies it also uses more resources
unnecessarily.

Another very important aspect of the algorithm is that while the waits between
checks are regular, they are also jittered in the time dimension by a user
specified amount.

The actual sequence of steps that the SRR algorithm follows is the following:

1. retrieves current price
2. calculates imbalance
3. if imbalance is above threshold initiates a rebalancing trade
4. waits jittered amount of time and refresh funds as the trade completes
5. repeats this sequence from step 1.

## Usage Guide

In Geronimo, everything is a node, and these nodes form a tree structure where
each node can have children. For example, an `Account` node represents a
connection to a CEX account, and users can create `Trader` nodes as its children
to facilitate autonomous trading on that account. The third type of node is
`Group`, which enables tree reorganization and creates passive fund partitions
called pockets. The final node type users need to understand is `User`, where
user-level settings are configured.

`Account` nodes can _hold_ funds, this is how funds are introduced to the tree.
`Group` and `Trader` nodes can _allocate_ funds from the available unallocated
holdings of their parents called _offer_. Allocated funds are then offered to
downstream nodes and can be further devided among the childrens. The GUI
collects and displays accumulated downstream statistics in each node's _info_
section, providing both detailed information and high-level oversight of the
tree's state.

Nodes can also have _parameters_, values that the user can set to configure the
system's expected behavior.

## Experience Contracts

The system must comply with well-defined behavioral contracts to be considered
correct.

### UI Consistency

The user interface connects only to the back-end and does not communicate with
any other parties involved in trading. This prevents the front-end from seeing a
different reality than the back-end. In other words, users can always be
confident that what they see on the interface reflects how the back-end sees the
world.

### Actuality

The front-end continuously monitors the health of its connection to the server
and reports this status in the bottom status bar. When the connection is
healthy, the status light is green; when broken, it's red. When the connection
is healthy, users can be confident that the displayed parameter and info values
are consistent with the back-end state. The system complies with this contract
by continuously updating all displayed topological, info, and parameter values,
and by reloading the complete display state after every reconnection.

### User Changes

Topological and parameter changes initiated by users are only updated in the UI
after the backend reports successful commitment to persistent storage and
resends the parameter as an update. When the UI displays the updated value,
users can be confident that the update will persist after a restart.

### Parameter History

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
