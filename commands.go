package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nsqio/go-nsq"
	"github.com/shamaton/msgpack"
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

	maxValue := ctx.Float64("max-value")
	radius := ctx.Float64("radius")
	distanceLimit := ctx.Float64("distance-limit")
	scanCount := ctx.Int("scans")
	buckets := ctx.Int("buckets")

	minScore := ctx.Float64("min-score")

	zline, err := NewZLine(origin, zeros, maxValue, 1, false)
	s := NewSession(cctx, 0, zline, lattice, radius, distanceLimit, minScore, scanCount, buckets)
	log.Println("Session scanning", scanCount, "origins")
	ch, err := s.Start()
	if err != nil {
		return err
	}

	count := 0
	running := true

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Instantiate a producer.
	topic := "scan-results"
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Fatal(err)
	}

	for running {
		select {
		case result, ok := <-ch:
			if !ok {
				log.Println("Channel closed. Stopping")
				producer.Stop()
				running = false
			} else {
				count++
				body, err := msgpack.Encode(result)
				err = producer.Publish(topic, body)
				if err != nil {
					log.Fatal(err)
				}
			}

		case <-sigChan:
			cancel()
			msg := fmt.Sprint("\nCanceled by user. count:", count)
			log.Fatal(msg)
		default:
		}
	}

	log.Println("Published", count, "points with a score >", s.minScore*100, "% at", s.ScansPerSec, "scans/sec in", s.TotalTime)

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

	minScore := .5 // 50% zeros hit

	ch := make(chan *Session)

	go func() {
		defer close(ch)
		for id, origin := range origins {
			zline, err := NewZLine(origin, zeros, maxValue, 1, false)
			if err != nil {
				return
			}

			ch <- NewSession(context.Background(), id, zline, lattice, radius, distanceLimit, minScore, scanCount, buckets)
		}
	}()

	return ch, nil
}
