package scan

import (
	"context"
	"log"
	"math"
	g "reticle/geom"
	"runtime"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
)

// Session is a distinct scan of random points within a radius from the ZLine
// origin.
type Session struct {
	ID            int64
	ZLine         g.ZLine
	Lattice       g.Lattice
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
func NewSession(id int64, zline g.ZLine, lattice g.Lattice, radius, distanceLimit, minScore float64, scansReq, bucketCount int) *Session {
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
		wg:            &sync.WaitGroup{},
		ctx:           cctx,
		cancel:        cancel,
	}

	s.wg.Add(s.ProcCount)
	return s
}

// RestoreSession rebuilds a Session from a deserialized Session from the message
// bus (basically the zeros values and lattice points are not there when
// serialized to the message bus)
func RestoreSession(s *Session) error {

	s.ProcCount = runtime.GOMAXPROCS(0)

	lattice, err := g.NewLattice(s.Lattice.LatticeType, s.Lattice.VertexType)
	if err != nil {
		return err
	}
	s.Lattice = lattice

	zeros := make([]g.Zeros, 0)
	for _, z := range s.ZLine.Zeros {
		zero, err := g.LoadZeros(z.ZeroType, s.ZLine.Limit, z.Scalar, z.Negatives)
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

// sessionFromCLI creates a session from CLI arguments and flags
func sessionFromCLI(ctx *cli.Context) (*Session, error) {

	var lt g.LatticeType
	lt, err := lt.GetLType(ctx.Args().Get(0))
	if err != nil {
		return nil, err
	}

	lattice, err := g.NewLattice(lt, g.Vertices)
	if err != nil {
		return nil, err
	}

	zeros := make([]g.ZeroType, 0)
	for _, zarg := range ctx.Args().Slice()[1:] {
		var zt g.ZeroType
		zt, err := zt.GetZType(zarg)
		if err != nil {
			return nil, err
		}

		zeros = append(zeros, zt)
	}

	origin := g.Vector2{
		X: ctx.Float64Slice("origin")[0],
		Y: ctx.Float64Slice("origin")[1],
	}

	maxValue := ctx.Float64("max-zero")
	radius := ctx.Float64("radius")
	distanceLimit := ctx.Float64("distance-limit")
	scanCount := ctx.Int("scans")
	buckets := ctx.Int("buckets")

	minScore := ctx.Float64("min-score")

	zline, err := g.NewZLine(origin, zeros, maxValue, 1, false)
	if err != nil {
		return nil, err
	}

	return NewSession(0, zline, lattice, radius, distanceLimit, minScore, scanCount, buckets), nil
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

func (s *Session) scanJob(id int, filtered []g.Vector2, resCh chan<- Result) {

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
			result := CreateResult(s.ID, s.BucketCount, origin, zero.ZeroType, zero.Count, hits)
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
