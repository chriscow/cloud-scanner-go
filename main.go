package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/joho/godotenv"
)

const bucketCount = 3600

func checkEnv() {
	godotenv.Load()

}

func main() {
	checkEnv()

	ips := getLocalIPs()
	fmt.Println(ips)

	lattice, err := NewLattice(Pinwheel, Vertices)
	if err != nil {
		log.Fatal(err)
	}

	primes, err := LoadZeros(Primes, 100, 1, false)
	if err != nil {
		log.Fatal(err)
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

	log.Println("Threads:", runtime.GOMAXPROCS(0), "calcs:", scanCount, "Elapsed", elapsed, scansPerSec, "scans/sec")

	cancel()
}
