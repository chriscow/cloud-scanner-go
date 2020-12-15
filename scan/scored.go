package scan

import (
	"context"
	"log"
	"time"
)

// ScoredResults keeps a sorted list of results ordered by score. Once the buffer
// contains `depth` results, we start popping the top results off and sent over
// the a channel given to us.
type ScoredResults struct {
	results resultHeap
	ctx     context.Context
	depth   int
	pub     chan<- []Result
	res     chan Result
}

// NewScoredResults creates and returns a ScoredResults instance
func NewScoredResults(ctx context.Context, depth int, publish chan<- []Result) *ScoredResults {
	sr := &ScoredResults{
		results: make([]Result, 0),
		pub:     publish,
		res:     make(chan Result),
		ctx:     ctx,
	}

	sr.start()

	return sr
}

func (sr *ScoredResults) start() {
	go func() {
		ticker := time.NewTicker(time.Second)
		running := true
		for running {
			select {
			case <-ticker.C:
				// every second, calc how many results are over our `depth`
				// value, make a slice, pop them off and publish them
				n := sr.results.Len()
				count := sr.depth - n
				if count > 0 {
					res := make([]Result, 0, count)
					for count > 0 {
						res = append(res, sr.results.Pop().(Result))
					}
					sr.pub <- res
				}
			case res := <-sr.res:
				sr.results.Push(res)
			case <-sr.ctx.Done():
				log.Println("[ScoredResults] Canceled: exiting")
				return
			}
		}
	}()
}

// Add adds the given result to our heap by sending through an internal channel
func (sr *ScoredResults) Add(res Result) {
	sr.res <- res
}
