package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// Session is a distinct scan of random points within a radius from the ZLine
// origin.
type Session struct {
	ID            int
	ZLine         *ZLine
	Lattice       *Lattice
	Radius        float64
	DistanceLimit float64
	BucketCount   int
	ScansPerSec   int
	TotalTime     time.Duration
	ProcCount     int
	ScansReq      int
	MinScore      float64
	wg            *sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewSession creates and initializes a new Session
func NewSession(id int, zline *ZLine, lattice *Lattice, radius, distanceLimit, minScore float64, scansReq, bucketCount int) *Session {
	cctx, cancel := context.WithCancel(context.Background())

	s := &Session{
		ID:            id,
		ZLine:         zline,
		Lattice:       lattice,
		Radius:        1,
		DistanceLimit: distanceLimit,
		BucketCount:   bucketCount,
		ProcCount:     runtime.GOMAXPROCS(0),
		MinScore:      minScore,
		ScansReq:      scansReq,
		ctx:           cctx,
		cancel:        cancel,
		wg:            &sync.WaitGroup{},
	}

	s.wg.Add(s.ProcCount)
	return s
}

// RestoreSession rebuilds a Session from a deserialized Session from the message
// queue (basically the zeros values and lattice points are not there)
func RestoreSession(s *Session) error {

	s.ProcCount = runtime.GOMAXPROCS(0)

	lattice, err := NewLattice(s.Lattice.LatticeType, s.Lattice.VertexType)
	if err != nil {
		return err
	}
	s.Lattice = lattice

	zeros := make([]*Zeros, 0)
	for _, z := range s.ZLine.Zeros {
		zero, err := LoadZeros(z.ZeroType, s.ZLine.Limit, z.Scalar, z.Negatives)
		if err != nil {
			return err
		}

		zeros = append(zeros, zero)
	}
	s.ZLine.Zeros = zeros
	s.wg = &sync.WaitGroup{}
	s.wg.Add(s.ProcCount)
	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cancel
	return nil
}

// createResult creates a regular `zeros hit` result and scores it on the
// percentage of zeros hit to total zeros
func (s *Session) createResult(origin Vector2, ztype ZeroType, zcount int, bh bucketHits) Result {
	if origin.X == 0 && origin.Y == 0 {
		msg := fmt.Sprint("[createResult] received 0,0 origin")
		log.Println(msg)
	}

	// trim off extranious decimal places since theta's precision is
	// dependent on the number of buckets.  x2 just in case
	places := math.Pow10(len(strconv.Itoa(s.BucketCount)) * 2)
	score := math.Round(float64(bh.Hits)/float64(zcount)*places) / places
	theta := math.Round(bh.Theta*places) / places

	return Result{
		SessionID:  s.ID,
		Origin:     origin,
		ZeroType:   ztype,
		ZerosCount: zcount,
		ZerosHit:   bh.Hits,
		BestTheta:  theta,
		BestBucket: bh.Bucket,
		Score:      score,
	}
}

// Start starts scanning using the session's parameters
func (s *Session) Start() (<-chan Result, error) {

	resCh := make(chan Result, s.ScansReq)

	maxZero := s.ZLine.MaxZeroVal()
	filtered := s.Lattice.Filter(s.ZLine.Origin, s.Radius, maxZero, s.DistanceLimit)

	start := time.Now()

	go func() {
		defer close(resCh)

		for i := 0; i < s.ProcCount; i++ {
			go s.scanJob(i, filtered, resCh)
		}

		s.wg.Wait()

		elapsed := time.Since(start)
		s.TotalTime = elapsed
		s.ScansPerSec = int(math.Round(float64(s.ScansReq) / elapsed.Seconds()))
	}()

	return resCh, nil
}

// Stop cancels a currently running scan
func (s *Session) Stop() {
	s.cancel()
}

func (s *Session) scanJob(id int, filtered []Vector2, resCh chan<- Result) {

	count := s.ScansReq / s.ProcCount
	origins := randOrigins(-s.Radius, s.Radius, s.ZLine.Origin, count)
	log.Println("[Job:", id, "] started scanning", count, "origins")

	// need the same origin for all zeros in the zline so we
	// can do a diff result
	for _, origin := range origins {

		zero := s.ZLine.Zeros[0]

		buckets := calculate(origin, filtered, zero.Values, s.Lattice.Parameters,
			s.DistanceLimit, s.BucketCount)

		best := getBestBuckets(buckets)
		for _, hits := range best {
			result := s.createResult(origin, zero.ZeroType, zero.Count, hits)
			if result.Score >= s.MinScore {
				resCh <- result
			}
		}

		// if len(s.ZLine.Zeros) == 2 {
		// 	// TODO: create a diff result
		// 	}
		// }

		// TODO: handle more zeros if there are any
		// if len(s.ZLine.Zeros) > 2 {
		// 	log.Fatal("TODO: handle remaining zeros or reject request")
		// }

		// check if we have been canceled
		select {
		case <-s.ctx.Done():
			break
		default:
		}
	}

	s.wg.Done()
	// log.Println("[Job:", id, "] done")
}
