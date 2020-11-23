package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

const bucketCount = 3600

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
	}

	// ZLine         *ZLine
	// Lattice       *Lattice
	// Radius        float64
	// DistanceLimit float64
	// BucketCount   int
	// scanCount     int

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
					Usage:     "scan a lattice at random points from the origin within a radius",
					ArgsUsage: "lattice zero1 [zero2]",
					Action:    scanRadiusCmd,
					Flags: []cli.Flag{

						&cli.BoolFlag{
							Name:  "service",
							Usage: "Run as a service",
						},

						&cli.Float64SliceFlag{
							Name:        "origin",
							DefaultText: "0 0",
							Value:       cli.NewFloat64Slice(0, 0),
						},

						&cli.Int64Flag{
							Name:  "radius",
							Value: 1,
						},

						&cli.Float64SliceFlag{
							Name:  "distance-limits",
							Usage: "One or more distance limits",
							Value: cli.NewFloat64Slice(1),
						},

						&cli.Int64Flag{
							Name:  "buckets",
							Value: 3600,
						},

						&cli.Int64Flag{
							Name:  "scans",
							Value: 1000,
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
