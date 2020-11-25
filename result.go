package main

import (
	"fmt"
	"sort"
)

// Result holds the data from a single scan and is serialized with MessagePack
type Result struct {
	SessionID     int
	Origin        Vector2
	ZeroType      ZeroType
	ZerosCount    int
	ZerosHit      int
	BestTheta     float64
	BestBucket    int
	ZeroIDs       []int
	AvgParity     float64
	LatticeParams interface{}
	Score         float64
}

func (r Result) String() string {
	return fmt.Sprint("[Result] session id:", r.SessionID,
		" score:", r.Score,
		" hits:", r.ZerosHit,
		" zeros:", r.ZerosCount,
		" theta:", r.BestTheta,
		" bucket:", r.BestBucket,
		" origin:", r.Origin, r.ZeroType)
}

// bucketHits holds the tally of hits from within a bucket
type bucketHits struct {
	Bucket int
	Hits   int
	Theta  float64
}

// String returns the string representation of a bucketHits struct
func (zh bucketHits) String() string {
	return fmt.Sprint("bucket:", zh.Bucket, " hits:", zh.Hits, " theta:", zh.Theta)
}

// countHits returns a sorted list of the count of hits in each bucket
func countHits(buckets [][]int) []bucketHits {
	hits := make([]bucketHits, len(buckets))
	degPerBucket := 360.0 / float64(len(buckets))

	for i, bucket := range buckets {
		var count int

		// the value in a bucket is either zero or one
		for _, hit := range bucket {
			count += hit
		}

		hits[i] = bucketHits{Bucket: i, Hits: count, Theta: float64(i) * degPerBucket}
	}

	sort.Slice(hits, func(a, b int) bool {
		return hits[a].Hits > hits[b].Hits
	})

	return hits
}

func getZerosHit(bucket []int, zeros []float64) []float64 {
	hits := make([]float64, 0)

	for i := range bucket {
		if bucket[i] > 0 {
			hits = append(hits, zeros[i])
		}
	}

	return zeros
}

// getResults finds the bucket(s) with the most hits and returns an array of
// bucketHits structs with the top hit counts
func getBestBuckets(buckets [][]int) []bucketHits {
	results := make([]bucketHits, 0)
	hits := countHits(buckets)

	best := hits[0]
	results = append(results, best)

	// We want to capture all results that are tied with the best
	for i := 1; i < len(hits); i++ {
		if hits[i].Hits == best.Hits {
			results = append(results, hits[i])
		}
	}

	return results
}
