package main

import (
	"testing"
)

func TestFitTo01(t *testing.T) {
	res := fitTo01(20, 10, 50)
	const want = 0.25
	if res != want {
		t.Errorf("Wrong result from `fitTo01`: %v, expected %v .", res, want)
	}
	res = fitTo01(60, 10, 50)
	const want2 = 1.25
	if res != want2 {
		t.Errorf("Wrong result from `fitTo01`: %v, expected %v .", res, want2)
	}
}

func TestFit01(t *testing.T) {
	res := fit01(.75, 10, 50)
	const want = 40
	if res != want {
		t.Errorf("Wrong result from `fit01`: %v, expected %v .", res, want)
	}
}

func TestClamp(t *testing.T) {
	res := clamp01(fitTo01(60, 10, 50))
	const want = 1
	if res != want {
		t.Errorf("Wrong result from `clamp`: %v, expected %v .", res, want)
	}
}
