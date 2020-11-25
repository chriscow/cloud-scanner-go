package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/urfave/cli/v2"
)

func scanRadiusCmd(ctx *cli.Context) error {

	if ctx.NArg() < 2 {
		return errors.New("Expected lattice and one or more zeros")
	}

	if ctx.Bool("service") {
		return scanRadiusSvc(ctx)
	}

	s, err := sessionFromCLI(ctx)
	if err != nil {
		return err
	}

	return startScan(s, nil)
}

// scanLatticeCmd generates scan-radius sessions and publishes them to the
// channel returned.  Each session contains a different origin such that all the
// scan sessions will completely cover the lattice.
func scanLatticeCmd(ctx *cli.Context) error {
	if ctx.NArg() < 2 {
		return errors.New("Expected lattice and one or more zeros")
	}

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Instantiate a producer.
	topic := "scan-session"
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Fatal(err)
	}

	// We create one session, thus only loading the lattice and zeros once
	// then just modify its ID and zline origin in the loop below
	s, err := sessionFromCLI(ctx)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	origins := s.Lattice.Partition(s.Radius)
	elapsed := time.Since(start)
	log.Println("Lattice partitioned in", elapsed.Seconds())

	wg := &sync.WaitGroup{}
	wg.Add(len(origins))

	start = time.Now()
	for id, origin := range origins {
		select {
		case <-sigChan:
			msg := fmt.Sprint("\nCanceled by user")
			log.Fatal(msg)
		default:
			go func() {

				s.ID = id
				s.ZLine.Origin = origin

				body, err := json.Marshal(s)
				if err != nil {
					log.Fatal(err)
				}

				err = producer.Publish(topic, body)
				if err != nil {
					log.Fatal(err)
				}

				wg.Done()
			}()
		}
	}

	wg.Wait()
	elapsed = time.Since(start)
	log.Println("Published", len(origins), "sessions in", elapsed.Seconds())

	return nil
}

func sessionFromCLI(ctx *cli.Context) (*Session, error) {

	var lt LatticeType
	lt, err := lt.GetLType(ctx.Args().Get(0))
	if err != nil {
		return nil, err
	}

	lattice, err := NewLattice(lt, Vertices)
	if err != nil {
		return nil, err
	}

	zeros := make([]ZeroType, 0)
	for _, zarg := range ctx.Args().Slice()[1:] {
		var zt ZeroType
		zt, err := zt.GetZType(zarg)
		if err != nil {
			return nil, err
		}

		zeros = append(zeros, zt)
	}

	origin := Vector2{
		X: ctx.Float64Slice("origin")[0],
		Y: ctx.Float64Slice("origin")[1],
	}

	maxValue := ctx.Float64("max-zero")
	radius := ctx.Float64("radius")
	distanceLimit := ctx.Float64("distance-limit")
	scanCount := ctx.Int("scans")
	buckets := ctx.Int("buckets")

	minScore := ctx.Float64("min-score")

	zline, err := NewZLine(origin, zeros, maxValue, 1, false)
	if err != nil {
		return nil, err
	}

	return NewSession(0, zline, lattice, radius, distanceLimit, minScore, scanCount, buckets), nil
}
