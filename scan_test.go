package main

import (
	"sync"
	"testing"
	"time"
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
	path := "../lattices/Pinwheel/Vertices/8192/points/"
	lattice := LoadLattice(Pinwheel, Vertices)

	path = "../zeros/primes.txt"
	zeros := LoadZ(path, 25)

	center := point{}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	calculate(center, 1, 1, lattice, zeros, wg)
	wg.Wait()
}

func TestStress(t *testing.T) {
	path := "../lattices/Pinwheel/Vertices/8192/points/"
	lattice := loadPoints(path)

	path = "../zeros/primes.txt"
	zeros := loadZeros(path, 25)

	cores := 8
	scansPerCore := 300

	center := point{x: 0, y: 0}
	radius := 1.0
	lattice = filterLattice(center, radius, lattice, zeros, 1)

	wg := new(sync.WaitGroup)
	wg.Add(cores)

	start := time.Now()
	for i := 0; i < cores; i++ {
		go calculate(center, radius, scansPerCore, lattice, zeros, wg)
	}

	wg.Wait()
	elapsed := time.Since(start)
	scansPerSec := float64(cores*scansPerCore) / elapsed.Seconds()
	// log.Println("elapsed", elapsed, scansPerSec, "scans/sec")
	if scansPerSec < 150 {
		t.Log("expected > 150 scans/sec but got elapsed", elapsed, scansPerSec, "scans/sec")
		t.Fail()
	}
}
