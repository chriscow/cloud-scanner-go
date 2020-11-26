package main

import (
	"context"
	"log"
	"math"
)

// ScoredResults keeps a sorted list of results ordered by score
type ScoredResults struct {
	results map[int][]Result
	resCh   <-chan Result
	ctx     context.Context
}

// NewScoredResults creates and returns a ScoredResults instance
func NewScoredResults(ctx context.Context, resCh <-chan Result) *ScoredResults {
	sr := &ScoredResults{
		results: make(map[int][]Result),
		resCh:   resCh,
		ctx:     ctx,
	}

	sr.start()

	return sr
}

// Scores returns all the result scores (rounded) that are currently stored
func (sr *ScoredResults) Scores() []int {
	scores := make([]int, len(sr.results))

	i := 0
	for key := range sr.results {
		scores[i] = key
		i++
	}

	return scores
}

// Results returns a list of Results that have the same score (rounded)
func (sr *ScoredResults) Results(score int) []Result {
	return sr.results[score]
}

func (sr *ScoredResults) start() {
	go func() {
		running := true
		for running {
			select {
			case res, ok := <-sr.resCh:
				if ok {
					sr.addResult(res)
				} else {
					// channel closed
					running = false
					log.Println("[ScoredResults] Channel closed: exiting")
				}
			case <-sr.ctx.Done():
				log.Println("[ScoredResults] Canceled: exiting")
				return
			}
		}
	}()
}

func (sr *ScoredResults) addResult(res Result) {
	// See if we have the score in the map already.
	// If not allocate one
	score := int(math.Round(res.Score))
	if _, ok := sr.results[score]; !ok {
		sr.results[score] = make([]Result, 0)
	}

	sr.results[score] = append(sr.results[score], res)
}

// func (s *ScoredResults) ScoreCount() int {

// }

// func (s *ScoredResults) ValuesCount() int {

// }
