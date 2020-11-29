package scan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/shamaton/msgpack"
	"github.com/urfave/cli/v2"
)

var (
	distances = []float64{.5, 1, 2, 4, 8, 16, 32, 64, math.MaxFloat64}
)

// Publish starts a scan based on the Session parameters and publishes
// the results to the NSQ message bus in the scan-radius-results topic
func Publish(ctx context.Context, topic string, s *Session) error {
	ch, err := s.Start()
	if err != nil {
		return err
	}

	count := 0
	running := true

	// Instantiate a producer.
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer("127.0.0.1:4150", config) // always produce to the local queue
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
					producer.Stop()
					return err
				}
			}

		case <-ctx.Done():
			producer.Stop()
			return ctx.Err()
		default:
		}
	}

	log.Println("Published", count, "points with a score >", s.MinScore*100, "% at", s.ScansPerSec, "scans/sec in", s.TotalTime)

	return nil
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
	cctx, _ := context.WithCancel(context.Background()) // actually unused
	s, err := sessionFromCLI(cctx, ctx)
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

				s.ID = start.Unix() + int64(id)
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
