package scanner

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
