package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/gocarina/gocsv"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

var (
	distances = []float64{.5, 1, 2, 4, 8, 16, 32, 64, math.MaxFloat64}
)

func checkEnv() {
	godotenv.Load()

}

func main() {

	app := &cli.App{
		Name:        "reticle",
		Usage:       "app usage",
		UsageText:   "app usage text",
		ArgsUsage:   "app argsusage",
		Description: "app description",
	}

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "Enable detailed output",
		},
		&cli.BoolFlag{
			Name:  "service",
			Usage: "Run as a service",
		},
	}

	// ZLine         *ZLine
	// Lattice       *Lattice
	// Radius        float64
	// DistanceLimit float64
	// BucketCount   int
	// scanCount     int

	scanFlags := []cli.Flag{

		&cli.Float64SliceFlag{
			Name:        "origin",
			DefaultText: "0 0",
			Value:       cli.NewFloat64Slice(0, 0),
		},

		&cli.Int64Flag{
			Name:  "radius",
			Value: 1,
		},

		&cli.Float64Flag{
			Name:  "max-value",
			Usage: "Limit the zeros so they don't exceed this value (scans are quicker)",
			Value: 100,
		},

		&cli.Float64Flag{
			Name:  "distance-limit",
			Usage: "Only consider hits within this distance from the zline",
			Value: 1,
		},

		&cli.IntFlag{
			Name:  "buckets",
			Value: 3600,
		},

		&cli.IntFlag{
			Name:  "scans",
			Value: 5000,
		},

		&cli.Float64Flag{
			Name:  "min-score",
			Usage: "Filter scores less than this float value (.3 == filter scores < 30%",
			Value: 0,
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:        "scan",
			Usage:       "Scans a lattice either as a service or directly via cli",
			UsageText:   "scan-usage-text",
			ArgsUsage:   "lattice zero1 [zero2...]",
			Description: "scan-description",

			Subcommands: []*cli.Command{
				{
					Name:      "radius",
					Usage:     "scan a lattice at random Points from the origin within a radius",
					ArgsUsage: "lattice zero1 [zero2]",
					Action:    scanRadiusCmd,
					Flags:     scanFlags,
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

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
