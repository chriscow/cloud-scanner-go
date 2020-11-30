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
}

// NewSession creates and initializes a new Session
func NewSession(id int64, zline g.ZLine, lattice g.Lattice, radius, distanceLimit, minScore float64, scansReq, bucketCount int) *Session {
	if id == 0 {
		id = time.Now().UnixNano()
	}

	if minScore == 0 {
		// if minScore is zero, we will publish every bucket so set a minimum
		// of 1 hit
		minScore = float64(1) / float64(zline.Zeros[0].Count)
	}

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
	}

	return s
}

// Restore rebuilds a Session from a deserialized Session from the message
// bus (basically the zeros values and lattice points are not there when
// serialized to the message bus)
func Restore(s *Session) error {

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
	return nil
}

// Start starts scanning using the session's parameters
func (s *Session) Start(ctx context.Context) (<-chan Result, error) {

	resCh := make(chan Result, s.ScansReq)

	maxZero := s.ZLine.MaxZeroVal()
	filtered := s.Lattice.Filter(s.ZLine.Origin, s.Radius, maxZero, s.DistanceLimit)

	start := time.Now()

	go func() {
		defer close(resCh)

		wg := &sync.WaitGroup{}
		wg.Add(s.ProcCount)

		for i := 0; i < s.ProcCount; i++ {
			go s.scanJob(ctx, wg, i, filtered, resCh)
		}

		wg.Wait()

		elapsed := time.Since(start)
		s.TotalTime = elapsed
		s.ScansPerSec = int(math.Round(float64(s.ScansReq) / elapsed.Seconds()))
	}()

	return resCh, nil
}

// scanJob generates random origins and scans them, publishing the results that
// meet the minimum score criteria. The number of random origins generated is
// determined by dividing the scans requested by the processor count, assuming
// scanJob will be called once per processor.
func (s *Session) scanJob(ctx context.Context, wg *sync.WaitGroup, procid int, filtered []g.Vector2, resCh chan<- Result) {

	count := s.ScansReq / s.ProcCount
	origins := randOrigins(-s.Radius, s.Radius, s.ZLine.Origin, count)
	log.Println("[Job:", procid, "] started scanning", count, "origins")

	// need the same origin for all zeros in the zline so we
	// can do a diff result
	for i, origin := range origins {

		zero := s.ZLine.Zeros[0]

		buckets := calculate(origin, filtered, zero.Values, s.Lattice.Parameters,
			s.DistanceLimit, s.BucketCount)

		best := getBestBuckets(buckets)
		for _, hits := range best {
			result := CreateResult(s.ID, procid, i, s.BucketCount, origin, zero.ZeroType, zero.Count, hits)
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
		case <-ctx.Done():
			wg.Done()
			return
		default:
		}
	}

	wg.Done()
	// log.Println("[Job:", id, "] done")
}

// sessionFromCLI creates a session from CLI arguments and flags
func sessionFromCLI(cctx context.Context, ctx *cli.Context) (*Session, error) {

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
