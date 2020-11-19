package main

import (
	"sync"
	"testing"
)

func TestCanLoadPinwheelVertices(t *testing.T) {
	lattice, err := LoadLattice(Pinwheel, Vertices)
	if err != nil {
		t.Fatal(err)
	}

	if lattice.LatticeType != Pinwheel {
		t.Log("expected lattice type == Pinwheel but was", lattice.LatticeType.String())
		t.Fail()
	}

	if lattice.VertexType != Vertices {
		t.Log("expexted vertex type == Vertices but was", lattice.VertexType.String())
		t.Fail()
	}
	if len(lattice.Points) != 277845 {
		t.Log("expected pinwheel verticies to have 277845 points but it had", len(lattice.Points))
		t.Fail()
	}
}

func TestAllAngles(t *testing.T) {

	lattice := Point{143.8583984375, -28.27120018005371}
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
	lattice, err := LoadLattice(Pinwheel, Vertices)
	if err != nil {
		t.Fatal(err)
	}
	zeros, err := LoadZeros(Primes, 100, 1, false)
	if err != nil {
		t.Fatal(err)
	}

	center := Point{}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	calculate(center, 1, 10, lattice.Points, zeros.Values, wg)
	wg.Wait()
}
