package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"
	"sync"

	"golang.org/x/sync/semaphore"
)

const (
	topic = "scan-results"
)

var (
	maxWorkers = runtime.GOMAXPROCS(0)
	sem        = semaphore.NewWeighted(int64(maxWorkers))
	distances  = []float64{.5, 1, 2, 4, 8, 16, 32, 64, math.MaxFloat64}
)

// Session is a ...
type Session struct {
	ID            int64
	ZLine         *ZLine
	Lattice       *Lattice
	Radius        float64
	DistanceLimit float64
	BucketCount   int
	scanCount     int
	ctx           context.Context
}

// Result holds the data from a single scan and is serialized with MessagePack
type Result struct {
	SessionID     int64
	Origin        Point
	ZeroType      ZeroType
	ZerosCount    int
	ZerosHit      int
	BestTheta     float64
	BestBucket    int
	AvgParity     float64
	LatticeParams interface{}
	Score         float64
}

func (r Result) String() string {
	return fmt.Sprint("session id:", r.SessionID, " origin:", r.Origin, r.ZeroType,
		" zeros:", r.ZerosCount, " hits:", r.ZerosHit, " theta:", r.BestTheta,
		" bucket:", r.BestBucket)
}

// Create and initialize a new Session
func Create(ctx context.Context, zline *ZLine, lattice *Lattice, radius, distanceLimit float64, scanCount int) *Session {
	return &Session{
		ZLine:         zline,
		Lattice:       lattice,
		Radius:        1,
		DistanceLimit: 1,
		scanCount:     scanCount,
		ctx:           ctx,
	}
}

// Start starts scanning using the session's parameters
func (s *Session) Start() <-chan Result {

	// threads := 8 // in case we are on the Mac

	// if runtime.GOOS != "darwin" {
	// 	cpu, err := ghw.CPU()
	// 	if err == nil {
	// 		threads = int(cpu.TotalThreads)
	// 	}
	// }

	resCh := make(chan Result, maxWorkers)
	center := s.ZLine.Origin
	radius := s.Radius

	maxZero := s.ZLine.MaxZeroVal()

	countPerProc := s.scanCount / maxWorkers

	// ptcount := len(s.Lattice.Points) // keep the pre-filtered point count for logging
	lattice := s.Lattice.Filter(center, radius, maxZero, s.DistanceLimit)

	log.Println("zeros:", len(s.ZLine.Zeros[0].Values), "maxZero:", maxZero, "lattice:", len(lattice))

	go func() {
		defer close(resCh)

		wg := &sync.WaitGroup{}
		wg.Add(maxWorkers)

		for i := 0; i < maxWorkers; i++ {
			// When maxWorkers goroutines are in flight,
			// Acquire blocks until one of the workers finishes.
			if err := sem.Acquire(s.ctx, 1); err != nil {
				log.Printf("Failed to acquire semaphore: %v", err)
				break
			}
			go func(i int) {
				defer sem.Release(1)
				log.Println("job", i)
				origins := randOrigins(-radius, radius, center, countPerProc)

				// need the same origin for all zeros
				for _, origin := range origins {
					zero := s.ZLine.Zeros[0]
					calculate(origin, lattice, zero, s.Lattice.Parameters)

					// TODO: if len(s.ZLine.Zeros) == 2 { do a diff result }

					select {
					case <-s.ctx.Done():
						return
					default:
					}
				}
				log.Println("job", i, "done")
				wg.Done()
			}(i)

			select {
			case <-s.ctx.Done():
				log.Println("canceled")
				return
			default:
			}
		}

		wg.Wait()
	}()

	return resCh
}

func randOrigins(min, max float64, center Point, count int) []Point {
	origins := make([]Point, count)
	for i := range origins {
		origins[i] = Point{
			X: min + rand.Float64()*(max-min) + center.X,
			Y: min + rand.Float64()*(max-min) + center.Y,
		}
	}
	return origins
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

func getBestBucket(buckets [][]int) (bestBucket, zerosHit int) {
	for i, bucket := range buckets {
		var sum int
		for _, hit := range bucket {
			sum += hit
		}

		if sum > zerosHit {
			zerosHit = sum
			bestBucket = i
		}
	}

	return bestBucket, zerosHit
}

// calculate a single result
func calculate(origin Point, lattice []Point, zeros *Zeros, latticeParams interface{}) Result {
	buckets := make([][]int, 3600)
	for i := range buckets {
		buckets[i] = make([]int, zeros.Count)
	}

	degPerBucket := 360.0 / float64(len(buckets))

	for _, point := range lattice {
		for i, zero := range zeros.Values {
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
			buckets[b1][i] = 1
			if b1 != b2 {
				buckets[b2][i] = 1
			}
		}
	}

	bestBucket, zerosHit := getBestBucket(buckets)
	bestTheta := float64(bestBucket) * degPerBucket

	return Result{
		Origin:        origin,
		ZeroType:      zeros.ZeroType,
		ZerosCount:    zeros.Count,
		ZerosHit:      zerosHit,
		BestTheta:     bestTheta,
		BestBucket:    bestBucket,
		LatticeParams: latticeParams,
	}
}
