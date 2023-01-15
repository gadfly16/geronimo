package main

import (
	"testing"
)

func TestGetBias(t *testing.T) {
	res := getBias(1, 0.2, 5)
	const want = 0.5
	if res != want {
		t.Errorf("Wrong result from `getBias` (1): %v, expected %v .", res, want)
	}
	res = getBias(0.1, 0.2, 5)
	const want2 = 1
	if res != want2 {
		t.Errorf("Wrong result from `getBias` (2): %v, expected %v .", res, want2)
	}
	res = getBias(0.4, 0.2, 3.2)
	const want3 = 0.75
	if res != want3 {
		t.Errorf("Wrong result from `getBias` (3): %v, expected %v .", res, want3)
	}
}

func TestGetAmount(t *testing.T) {
	res := getAmount(1, 0.2, 5, 75, 25)
	const want1 = -25
	if res != want1 {
		t.Errorf("Wrong result from `getAmount` (1): %v, expected %v .", res, want1)
	}
	res = getAmount(0.4, 0.2, 3.2, 0, 100)
	const want2 = 187.5
	if res != want2 {
		t.Errorf("Wrong result from `getAmount` (2): %v, expected %v .", res, want2)
	}
}

func TestBrokerPrepare(t *testing.T) {
	bro := &broker{
		base:      0,
		quote:     100,
		highLimit: 3.2,
		lowLimit:  0.2,
		delta:     0.04,
		offset:    0.01,
	}
	lastOrd := &order{price: 0.5}
	ord := order{midPrice: 0.404}
	ord.prepareTrade(bro, lastOrd)
	res := ord.amount
	const want1 = 187.5
	if res != want1 {
		t.Errorf("Wrong result from `broker.prapare` (1): %v, expected %v .", res, want1)
	}
	lastOrd = &order{price: 0.402}
	ord = order{midPrice: 0.404}
	ord.prepareTrade(bro, lastOrd)
	res = ord.amount
	const want2 = 0
	if res != want2 {
		t.Errorf("Wrong result from `broker.prapare` (1): %v, expected %v .", res, want2)
	}

}
