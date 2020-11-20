package scanner

import (
	"log"
	"math"
	"math/rand"
	"reticle/geom"
	"runtime"
	"sync"
	"time"

	"github.com/jaypipes/ghw"
)

// Session is a ...
type Session struct {
	ZLine         *geom.ZLine
	Lattice       *geom.Lattice
	Radius        float64
	DistanceLimit float64
}

// Result holds the data from a single scan and is serialized with MessagePack
type Result struct {
	ZLine         *geom.ZLine
	BestTheta     float64
	ZerosCount    int
	ZerosHit      int
	AvgParity     float64
	LatticeParams interface{}
	Score         float64 `msgpack:"ignore"`
}

// Start starts scanning using the session's parameters
func (s *Session) Start(scansPerThread int) {

	threads := 8 // in case we are on the Mac

	if runtime.GOOS != "darwin" {
		cpu, err := ghw.CPU()
		if err == nil {
			threads = int(cpu.TotalThreads)
		}
	}

	center := s.ZLine.Origin
	radius := s.Radius

	maxZero := s.ZLine.MaxZeroVal()

	ptcount := len(s.Lattice.Points) // keep the pre-filtered point count for logging
	points := s.Lattice.Filter(center, radius, maxZero, s.DistanceLimit)

	wg := new(sync.WaitGroup)
	wg.Add(threads)

	start := time.Now()
	for i := 0; i < threads/len(s.ZLine.Zeros)+1; i++ {
		for _, zero := range s.ZLine.Zeros {
			go calculate(center, radius, scansPerThread, points, zero.Values, wg)
		}
	}

	wg.Wait()
	elapsed := time.Since(start)
	scansPerSec := float64(threads*scansPerThread) / elapsed.Seconds()

	log.Println("Lattice:", s.Lattice.LatticeType, s.Lattice.VertexType, "points:", ptcount, "filtered:", len(points))
	for _, zeros := range s.ZLine.Zeros {
		log.Println("ZLine:", zeros.ZeroType, "zeros count:", len(zeros.Values), "max zero:", maxZero)
	}
	log.Println("Threads:", threads, "Elapsed", elapsed, scansPerSec, "scans/sec")
}

func randOrigins(min, max float64, center geom.Point, count int) []geom.Point {
	res := make([]geom.Point, count)

	for i := range res {
		res[i] = geom.Point{
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

func allAngles(lattice, origin geom.Point, zero float64) (theta1, theta2 float64) {
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

func calculate(center geom.Point, radius float64, scans int, lattice []geom.Point, zeros []float64, wg *sync.WaitGroup) {
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
