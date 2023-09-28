package main

type accountMsg struct{}
type brokerMsg struct{}

var glRunningAccounts = map[int64](chan accountMsg){}
var runningBrokers = map[int64](chan brokerMsg){}
