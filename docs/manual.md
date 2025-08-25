# Geronimo Manual

This is the de facto documentation of Geronimo, which is a simple and reliable
trader application concentrating on the Stochastic Proportional Reblancer
algorithm.

[[_TOC_]]

## Introduction

Geronimo designed to run on a RasPI connected to a home grade router. This is a
very cost effective way to both provide decent availability and decent privacy
and security guarantees for the user. It connects to the user's CEX accounts and
facilitates trades conforming to the user's settings of the SPR algorithm. The
RasPi connected to the household router executes the back-end code and the user
accesses a web interface from a phone or desktop connected to the router. With
DDNS the appliance can be configured to be accessible from the outside world
too.

### The SPR Algorithm

Most trader bots allow the user to define, backtest and run custom algorithms.
Geronimo takes a different approach and tries to concentrate on one single
trading strategy called Stochastic Proportional Rebalancer. This focus allows us
to provide a user experience that we think is unique and can be desirable for
some people. It balances ease of use, ease of mind, calculable risks and costs
with competitive yields.

## Build Instructions

### Dependencies

You need a working Go environment and the following dependencies to produce the
back-end executable and the distributable version of the GUI:

```
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
