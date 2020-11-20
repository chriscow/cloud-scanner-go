package main

import (
	"testing"
)

func TestAllAngles(t *testing.T) {

	lattice := Point{X: 143.8583984375, Y: -28.27120018005371}
	zero := 13.0

	center := Point{}

	theta1, theta2 := allAngles(lattice, center, zero)

	if theta1 != 73.79472632488876 {
		t.Log(lattice, center, zero)
		t.Log("theta1 expected to be 73.79472632488876 but was", theta1)
		t.Fail()
	}

	if theta2 != 263.9689917699393 {
		t.Log(lattice, center, zero)
		t.Log("theta2 expected to be 263.9689917699393 but was", theta2)
		t.Fail()
	}
}

func TestCalc(t *testing.T) {
	lattice, err := NewLattice(Pinwheel, Vertices)
	if err != nil {
		t.Fatal(err)
	}
	zeros, err := LoadZeros(Primes, 100, 1, false)
	if err != nil {
		t.Fatal(err)
	}

	center := Point{}

	res := calculate(center, lattice.Points, zeros, nil)
	t.Log(res)
	t.Log("need to test result")
	t.Fail()
}
