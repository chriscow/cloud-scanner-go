package main

import (
	"log"
	"testing"
)

func TestAllAngles(t *testing.T) {

	lattice := Vector2{X: 0, Y: 0}
	zero := 13.0

	origin := Vector2{-0.8804702, -0.0348327}
	limit := 1
	theta1, theta2 := allAngles(lattice, origin, zero, limit)

	if theta1 != 73.79472632488876 {
		t.Log(lattice, origin, zero)
		t.Log("theta1 expected to be 73.79472632488876 but was", theta1)
		t.Fail()
	}

	if theta2 != 263.9689917699393 {
		t.Log(lattice, origin, zero)
		t.Log("theta2 expected to be 263.9689917699393 but was", theta2)
		t.Fail()
	}
}

func TestCalc(t *testing.T) {
	lattice, _ := NewLattice(Pinwheel, Vertices)
	zeros, _ := LoadZeros(Primes, 100, 1, false)
	origin := Vector2{-0.4297824, -0.9473751}
	log.Println(len(zeros.Values), "zeros", zeros.Values[0], zeros.Values[len(zeros.Values)-1])

	result := calculate(origin, lattice.Points, zeros, nil, 1, 3600)
}

// func TestPerf(t *testing.T) {
// 	checkEnv()

// 	ips := getLocalIPs()
// 	t.Log(ips)

// 	lattice, err := NewLattice(Pinwheel, Vertices)
// 	if err != nil {
// 		t.Log(err)
// 		t.Fail()
// 	}

// 	primes, err := LoadZeros(Primes, 100, 1, false)
// 	if err != nil {
// 		t.Log(err)
// 		t.Fail()
// 	}

// 	zeros := []*Zeros{primes}

// 	zline := &ZLine{
// 		Origin: Vector2{X: 0, Y: 0},
// 		Angle:  0,
// 		Zeros:  zeros,
// 	}

// 	ctx := context.Background()
// 	cctx, cancel := context.WithCancel(ctx)

// 	scanCount := 1000 * maxWorkers

// 	s := NewSession(cctx, zline, lattice, 1, 1, 100)

// 	start := time.Now()
// 	ch, err := s.Start()
// 	if err != nil {
// 		t.Log("Error starting scan")
// 		t.Log(err)
// 		t.Fail()
// 	}

// 	var count int
// 	for {
// 		select {
// 		case _ = <-ch:
// 			count++
// 		default:
// 		}
// 	}

// 	elapsed := time.Since(start)
// 	scansPerSec := float64(scanCount) / elapsed.Seconds()

// 	t.Log("Threads:", runtime.GOMAXPROCS(0), "calcs:", scanCount, "Elapsed", elapsed, scansPerSec, "scans/sec")

// 	cancel()
// }
