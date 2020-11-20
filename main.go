package main

import (
	"context"
	"fmt"
	"log"

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

	primes, err := LoadZeros(Primes, 50, 1, false)
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

	s := Create(cctx, zline, lattice, 1, 1)

	s.Start()
	
	startConsumer()

	cancel()
}
