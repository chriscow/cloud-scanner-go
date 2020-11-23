package main

import (
	"context"
	"errors"
	"log"
)

// Session is a ...
type Session struct {
	ID             int64
	ZLine          *ZLine
	Lattice        *Lattice
	Radius         float64
	DistanceLimits []float64
	BucketCount    int
	scanCount      int
	ctx            context.Context
}

// NewSession creates and initializes a new Session
func NewSession(ctx context.Context, zline *ZLine, lattice *Lattice, radius float64, distanceLimits []float64, scanCount int) *Session {
	return &Session{
		ZLine:          zline,
		Lattice:        lattice,
		Radius:         1,
		DistanceLimits: distanceLimits,
		scanCount:      scanCount,
		ctx:            ctx,
	}
}

// Start starts scanning using the session's parameters
func (s *Session) Start() (<-chan Result, error) {

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
	if len(s.DistanceLimits) > 1 {
		return nil, errors.New("Only a single distance limit is currently supported")
	}

	limit := s.DistanceLimits[0]

	lattice := s.Lattice.Filter(center, radius, maxZero, limit)

	log.Println("zeros:", len(s.ZLine.Zeros[0].Values), "maxZero:", maxZero, "lattice:", len(lattice))

	go func() {
		defer close(resCh)

		for {
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
						result1 := calculate(origin, lattice, zero, s.Lattice.Parameters, limit)
						resCh <- result1

						if len(s.ZLine.Zeros) == 2 {
							result2 := calculate(origin, lattice, zero, s.Lattice.Parameters, limit)
							resCh <- result2

							// TODO: create a diff result
							log.Printf("TODO: do a diff result")
						}

						// TODO: handle more zeros if there are any
						if len(s.ZLine.Zeros) > 2 {
							log.Printf("TODO: handle remaining zeros or reject request")
						}

						select {
						case <-s.ctx.Done():
							return
						default:
						}
					}
					log.Println("job", i, "done")
				}(i)

				select {
				case <-s.ctx.Done():
					log.Println("canceled")
					return
				default:
				}
			}
		}
	}()

	return resCh, nil
}
