package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"
	"strings"
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

func sampleScan() {
	lattice, err := LoadLattice(Pinwheel, Vertices)
	if err != nil {
		log.Fatal(err)
	}

	primes, err := LoadZeros(Primes, 100, 1, false)
	if err != nil {
		log.Fatal(err)
	}

	threads := 8

	log.Println("os", runtime.GOOS)
	if runtime.GOOS != "darwin" {
		threads = 0
		cpu, err := ghw.CPU()
		if err != nil {
			fmt.Printf("Error getting CPU info: %v", err)
		} else {

			for _, proc := range cpu.Processors {
				fmt.Printf(" %v\n", proc)
				for _, core := range proc.Cores {
					fmt.Printf("  %v\n", core)
				}
				if len(proc.Capabilities) > 0 {
					// pretty-print the (large) block of capability strings into rows
					// of 6 capability strings
					rows := int(math.Ceil(float64(len(proc.Capabilities)) / float64(6)))
					for row := 1; row < rows; row = row + 1 {
						rowStart := (row * 6) - 1
						rowEnd := int(math.Min(float64(rowStart+6), float64(len(proc.Capabilities))))
						rowElems := proc.Capabilities[rowStart:rowEnd]
						capStr := strings.Join(rowElems, " ")
						if row == 1 {
							fmt.Printf("  capabilities: [%s\n", capStr)
						} else if rowEnd < len(proc.Capabilities) {
							fmt.Printf("                 %s\n", capStr)
						} else {
							fmt.Printf("                 %s]\n", capStr)
						}
					}
				}
			}

			threads = int(cpu.TotalThreads)
			log.Println("total cores:", cpu.TotalCores, "total threads:", cpu.TotalThreads)
		}
	}

	log.Println("threads:", threads)
	scansPerThread := 100

	center := Point{X: 0, Y: 0}
	radius := 1.0
	lattice.Filter(center, radius, primes.Values, 1)

	wg := new(sync.WaitGroup)
	wg.Add(threads)

	start := time.Now()
	for i := 0; i < threads; i++ {
		go calculate(center, radius, scansPerThread, lattice.Points, primes.Values, wg)
	}

	wg.Wait()
	elapsed := time.Since(start)
	scansPerSec := float64(threads*scansPerThread) / elapsed.Seconds()
	log.Println("elapsed", elapsed, scansPerSec, "scans/sec")
}
