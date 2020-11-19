package main

import (
	"log"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/jaypipes/ghw"
)

func randOrigins(min, max float64, center Point, count int) []Point {
	res := make([]Point, count)

	for i := range res {
		res[i] = Point{
			X: min + rand.Float64()*(max-min) + center.X,
			Y: min + rand.Float64()*(max-min) + center.Y,
		}
	}
	return res
}

func wrapDegrees(deg float64) float64 {
	for deg > 360 {
		deg -= 360
	}

	for deg < 0 {
		deg += 360
	}

	return deg
}

func allAngles(lattice, origin Point, zero float64) (theta1, theta2 float64) {
	xSq := (lattice.X - origin.X) * (lattice.X - origin.X)
	ySq := (lattice.Y - origin.Y) * (lattice.Y - origin.Y)
	zSq := zero * zero

	distance := math.Sqrt(xSq + ySq - zSq)

	if math.IsNaN(distance) {
		return math.NaN(), math.NaN()
	}

	rad2deg := 180 / math.Pi
	theta1 = wrapDegrees(rad2deg * 2 * math.Atan2(lattice.Y-origin.Y+distance, lattice.X-origin.X+zero))
	theta2 = wrapDegrees(rad2deg * 2 * math.Atan2(lattice.Y-origin.Y-distance, lattice.X-origin.X+zero))

	return
}

func calculate(center Point, radius float64, scans int, lattice []Point, zeros []float64, wg *sync.WaitGroup) {
	buckets := make([]int, 3600, 3600)
	degPerBucket := 360.0 / float64(len(buckets))

	origins := randOrigins(-radius, radius, center, scans)

	for _, point := range lattice {
		for _, zero := range zeros {
			for _, origin := range origins {
				theta1, theta2 := allAngles(point, origin, zero)

				if math.IsNaN(theta1) || math.IsNaN(theta2) {
					continue
				}

				b1 := int(math.Floor(theta1 / degPerBucket))
				b2 := int(math.Floor(theta2 / degPerBucket))
				// log.Println(theta1, theta2, b1, b2)

				if b1 >= len(buckets) || b2 >= len(buckets) || b1 < 0 || b2 < 0 {
					log.Fatalln("bucket", b1, "out of range:", len(buckets), lattice, origin, zero, theta1, theta2)
				}
				buckets[b1]++
				if b1 != b2 {
					buckets[b2]++
				}
			}
		}
	}

	wg.Done()
}

func sampleScan(zeroCount int, scansPerThread int) {
	lattice, err := LoadLattice(Pinwheel, Vertices)
	if err != nil {
		log.Fatal(err)
	}

	zeros, err := LoadZeros(Primes, 100, 1, false)
	if err != nil {
		log.Fatal(err)
	}

	threads := 8 // in case we are on the Mac

	if runtime.GOOS != "darwin" {
		cpu, err := ghw.CPU()
		if err == nil {
			threads = int(cpu.TotalThreads)
		}
	}

	center := Point{X: 0, Y: 0}
	radius := 1.0

	ptcount := len(lattice.Points)
	lattice.Filter(center, radius, zeros.Values[:zeroCount], 1)

	wg := new(sync.WaitGroup)
	wg.Add(threads)

	start := time.Now()
	for i := 0; i < threads; i++ {
		go calculate(center, radius, scansPerThread, lattice.Points, zeros.Values[:zeroCount], wg)
	}

	wg.Wait()
	elapsed := time.Since(start)
	scansPerSec := float64(threads*scansPerThread) / elapsed.Seconds()

	log.Println("Lattice:", lattice.LatticeType, lattice.VertexType, "points:", ptcount, "filtered:", len(lattice.Points))
	log.Println("ZLine:", zeros.ZeroType, "zeros:", len(zeros.Values), "limited to:", zeroCount)
	log.Println("Threads:", threads, "Elapsed", elapsed, scansPerSec, "scans/sec")
}
