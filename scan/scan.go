package scan

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/nsqio/go-nsq"
	"github.com/shamaton/msgpack"
)

var (
	distances = []float64{.5, 1, 2, 4, 8, 16, 32, 64, math.MaxFloat64}
)

// ScanAndPublish starts a scan based on the Session parameters and publishes
// the results to the NSQ message bus in the scan-radius-results topic
func ScanAndPublish(ctx context.Context, s *Session) error {
	ch, err := s.Start()
	if err != nil {
		return err
	}

	count := 0
	running := true

	// Instantiate a producer.
	topic := "scan-radius-results"
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
				s.Stop()
				running = false
			} else {
				count++
				body, err := msgpack.Encode(result)
				err = producer.Publish(topic, body)
				if err != nil {
					s.Stop()
					producer.Stop()
					log.Fatal(err)
				}
			}

		case <-ctx.Done():
			producer.Stop()
			s.Stop()
			msg := fmt.Sprint("\nCanceled by user. count:", count)
			log.Fatal(msg)
		default:
		}
	}

	log.Println("Published", count, "points with a score >", s.MinScore*100, "% at", s.ScansPerSec, "scans/sec in", s.TotalTime)

	return nil
}
