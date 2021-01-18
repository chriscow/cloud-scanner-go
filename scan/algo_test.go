package scan

import (
	"log"
	"math"
	"os"
	"github.com/chriscow/cloud-scanner-go/geom"
	"testing"
)

func TestAllAngles(t *testing.T) {

	lattice := geom.Vector2{X: 123.45, Y: 456.789}
	zero := 13.0

	origin := geom.Vector2{X: -0.8804702, Y: -0.0348327}
	limit := math.MaxFloat64
	theta1, theta2 := allAngles(lattice, origin, zero, limit)

	if theta1 != 163.2015290584845 {
		t.Log(lattice, origin, zero)
		t.Log("theta1 expected to be 163.2015290584845 but was", theta1)
		t.Fail()
	}

	if theta2 != 346.3484439376577 {
		t.Log(lattice, origin, zero)
		t.Log("theta2 expected to be 346.3484439376577 but was", theta2)
		t.Fail()
	}
}

func TestCalc(t *testing.T) {
	lattice, _ := geom.NewLattice(geom.Pinwheel, geom.Vertices)

	zeros := geom.Zeros{
		ZeroType:  geom.Primes,
		Scalar:    1,
		Negatives: false,
	}
	err := geom.LoadZeros(&zeros, 100)
	if err != nil {
		t.Log("LoadZeros", err)
		t.Fail()
	}

	origin := geom.Vector2{X: -0.4297824, Y: -0.9473751}
	log.Println(len(zeros.Values), "zeros", zeros.Values[0], zeros.Values[len(zeros.Values)-1])

	buckets := 3600
	result := calculate(origin, lattice.Points, zeros.Values, nil, 1, buckets)
	if len(result) != buckets {
		t.Log("length of result should match bucket count", buckets, len(result))
		t.Fail()
	}

	for i := range result {
		if len(result[i]) != zeros.Count {
			t.Log("each bucket should be the same length as zeros count", len(result[i]), zeros.Count)
			t.Fail()
		}
	}

	// TODO: MORE CHECKING HERE
	t.Log("Insufficient verification in this test")
	t.Fail()
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
