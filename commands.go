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
	origin := Point{
		X: ctx.Float64Slice("origin")[0],
		Y: ctx.Float64Slice("origin")[1],
	}

	maxValue := float64(100)
	radius := ctx.Float64("radius")
	distanceLimit := ctx.Float64Slice("distance-limits")
	scanCount := int(ctx.Int64("scans"))

	zline, err := NewZLine(origin, zeros, maxValue, 1, false)
	s := NewSession(cctx, zline, lattice, radius, distanceLimit, scanCount)

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

func scanRadiusSvc(ctx *cli.Context) error {
	handler := scanResultHandler{}
	return startConsumer(handler)
}
