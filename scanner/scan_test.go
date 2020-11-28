package scanner

import (
	"fmt"
	"log"
	"os"

	"github.com/gocarina/gocsv"
)

// the `top1000` csv file from Unity Reticle
// x,y,limit,theta,hits,ztype,zcount,zscale,score,params
type scanTestArg struct {
	X          float64 `csv:"x"`
	Y          float64 `csv:"y"`
	Limit      float64 `csv:"limit"`
	NumBuckets int     `csv:"buckets"`
	Theta      float64 `csv:"theta"`
	Hits       int     `csv:"hits"`
	ZeroType   string  `csv:"ztype"`
	ZeroCount  int     `csv:"zcount"`
	Scalar     float64 `csv:"zscale"`
	Score      float64 `csv:"score"`
}

func (s scanTestArg) String() string {
	return fmt.Sprint("origin:", s.X, s.Y, " limit:", s.Limit, " buckets:", s.NumBuckets,
		" theta:", s.Theta, " hits:", s.Hits, " ztype:", s.ZeroType, " zcount:", s.ZeroCount,
		" scalar:", s.Scalar, " score:", s.Score)
}

func compareTest() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	path := cwd + "/data/test/pinwheel-hits-test.csv"
	testFile, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer testFile.Close()

	testargs := []*scanTestArg{}
	if err := gocsv.UnmarshalFile(testFile, &testargs); err != nil {
		panic(err)
	}

	lattice, _ := NewLattice(Pinwheel, Vertices)
	zeros, _ := LoadZeros(Primes, 100, 1, false)

	log.Println(len(zeros.Values), "zeros", zeros.Values[0], zeros.Values[len(zeros.Values)-1])

	maxZero := zeros.Values[len(zeros.Values)-1]
	points := lattice.Filter(Vector2{}, 1, maxZero, testargs[0].Limit)

	for line, arg := range testargs {
		origin := Vector2{X: arg.X, Y: arg.Y}
		buckets := calculate(origin, points, zeros.Values, nil, arg.Limit, arg.NumBuckets)
		best := getBestBuckets(buckets)
		if len(best) > 1 {
			log.Println("multiple results:", len(best))
			for _, result := range best {
				log.Println("bucket: ", result.Bucket, " hits: ", result.Hits, " theta:", result.Theta)
			}
		}

		for _, result := range best {
			if result.Hits != arg.Hits || result.Theta != arg.Theta {
				log.Println("mismatch on line", line, "at", origin, "bucket", arg.Theta)
				log.Println("\tgolang", result)
				log.Println("\treticle", arg)
			}
		}
	}
}
