package main

import (
	"log"
	"reticle/geom"
	"reticle/scanner"

	"github.com/joho/godotenv"
)

const bucketCount = 3600

func checkEnv() {
	godotenv.Load()

}

func main() {
	checkEnv()

	lattice, err := geom.NewLattice(geom.Pinwheel, geom.Vertices)
	if err != nil {
		log.Fatal(err)
	}

	primes, err := geom.LoadZeros(geom.Primes, 50, 1, false)
	if err != nil {
		log.Fatal(err)
	}

	zeros := []*geom.Zeros{primes}

	zline := &geom.ZLine{
		Origin: geom.Point{X: 0, Y: 0},
		Angle:  0,
		Zeros:  zeros,
	}

	session := scanner.Session{
		ZLine:         zline,
		Lattice:       lattice,
		Radius:        1,
		DistanceLimit: 1,
	}

	session.Start(100)
}
