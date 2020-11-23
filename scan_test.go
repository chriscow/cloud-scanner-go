package main

import (
	"context"
	"time"
	"runtime"
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

func TestPerf(t *testing.T) {
	checkEnv()

	ips := getLocalIPs()
	t.Log(ips)

	lattice, err := NewLattice(Pinwheel, Vertices)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	primes, err := LoadZeros(Primes, 100, 1, false)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	zeros := []*Zeros{primes}

	zline := &ZLine{
		Origin: Point{X: 0, Y: 0},
		Angle:  0,
		Zeros:  zeros,
	}

	ctx := context.Background()
	cctx, cancel := context.WithCancel(ctx)

	scanCount := 1000 * maxWorkers

	s := Create(cctx, zline, lattice, 1, 1, scanCount)

	start := time.Now()
	ch := s.Start()

	<-ch

	elapsed := time.Since(start)
	scansPerSec := float64(scanCount) / elapsed.Seconds()

	t.Log("Threads:", runtime.GOMAXPROCS(0), "calcs:", scanCount, "Elapsed", elapsed, scansPerSec, "scans/sec")

	cancel()
}
