package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/urfave/cli/v2"
)

func scanRadiusCmd(ctx *cli.Context) error {

	if ctx.NArg() < 2 {
		return errors.New("Expected lattice and one or more zeros")
	}

	if ctx.Bool("service") {
		return scanRadiusSvc(ctx)
	}

	cctx, cancel := context.WithCancel(context.Background())

	var lt LatticeType
	lt, err := lt.GetLType(ctx.Args().Get(0))
	if err != nil {
		log.Fatal(err)
	}

	lattice, err := NewLattice(lt, Vertices)
	if err != nil {
		log.Fatal(err)
	}

	zeros := make([]ZeroType, 0)
	for _, zarg := range ctx.Args().Slice()[1:] {
		var zt ZeroType
		zt, err := zt.GetZType(zarg)
		if err != nil {
			log.Fatal(err)
		}

		zeros = append(zeros, zt)
	}
	origin := Vector2{
		X: ctx.Float64Slice("origin")[0],
		Y: ctx.Float64Slice("origin")[1],
	}

	maxValue := float64(100)
	radius := ctx.Float64("radius")
	distanceLimit := ctx.Float64("distance-limit")
	scanCount := ctx.Int("scans")
	buckets := ctx.Int("buckets")

	zline, err := NewZLine(origin, zeros, maxValue, 1, false)
	s := NewSession(cctx, zline, lattice, radius, distanceLimit, scanCount, buckets)

	ch, err := s.Start()
	if err != nil {
		return err
	}

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	count := 0
	for {

		select {
		case result := <-ch:
			count++
			if result.Origin.X == 0 && result.Origin.Y == 0 {
				return errors.New(strconv.Itoa(count))
			}
		case <-sigChan:
			cancel()
			msg := fmt.Sprint("canceled by user. count:", count)
			log.Fatal(msg)
		default:
		}
	}

	return nil
}

// scanLatticeCmd generates scan-radius sessions and publishes them to the
// channel returned.  Each session contains a different origin such that all the
// scan sessions will completely cover the lattice.
func scanLatticeCmd(ctx *cli.Context) (<-chan *Session, error) {
	if ctx.NArg() < 2 {
		return nil, errors.New("Expected lattice and one or more zeros")
	}

	var lt LatticeType
	lt, err := lt.GetLType(ctx.Args().Get(0))
	if err != nil {
		log.Fatal(err)
	}

	lattice, err := NewLattice(lt, Vertices)
	if err != nil {
		log.Fatal(err)
	}

	zeros := make([]ZeroType, 0)
	for _, zarg := range ctx.Args().Slice()[1:] {
		var zt ZeroType
		zt, err := zt.GetZType(zarg)
		if err != nil {
			log.Fatal(err)
		}

		zeros = append(zeros, zt)
	}

	maxValue := float64(100)
	radius := ctx.Float64("radius")
	distanceLimit := ctx.Float64("distance-limit")
	scanCount := ctx.Int("scans")
	buckets := ctx.Int("buckets")

	origins := lattice.Partition(radius)

	ch := make(chan *Session)

	go func() {
		defer close(ch)
		for _, origin := range origins {
			zline, err := NewZLine(origin, zeros, maxValue, 1, false)
			if err != nil {
				return
			}

			ch <- NewSession(context.Background(), zline, lattice, radius, distanceLimit, scanCount, buckets)
		}
	}()

	return ch, nil
}
