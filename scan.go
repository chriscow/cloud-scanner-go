package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nsqio/go-nsq"
	"github.com/shamaton/msgpack"
)

// session is the deserialized session request.
func scanAndPublish(s *Session, sigChan <-chan os.Signal) error {
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

		case <-sigChan:
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
