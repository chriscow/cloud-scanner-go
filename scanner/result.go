package scanner

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
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

// CreateResult creates a regular `zeros hit` result and scores it on the
// percentage of zeros hit to total zeros
func CreateResult(id, bucketCount int, origin Vector2, ztype ZeroType, zcount int, bh bucketHits) Result {
	if origin.X == 0 && origin.Y == 0 {
		msg := fmt.Sprint("[createResult] received 0,0 origin")
		log.Println(msg)
	}

	// trim off extranious decimal places since theta's precision is
	// dependent on the number of buckets.  x2 just in case
	places := math.Pow10(len(strconv.Itoa(bucketCount)) * 2)
	score := math.Round(float64(bh.Hits)/float64(zcount)*places) / places
	theta := math.Round(bh.Theta*places) / places

	return Result{
		SessionID:  id,
		Origin:     origin,
		ZeroType:   ztype,
		ZerosCount: zcount,
		ZerosHit:   bh.Hits,
		BestTheta:  theta,
		BestBucket: bh.Bucket,
		Score:      score,
	}
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