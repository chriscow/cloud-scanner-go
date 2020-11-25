package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/shamaton/msgpack"
)

func startScan(s *Session, msg *nsq.Message) error {
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
	topic := "scan-radius-results"
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Fatal(err)
	}

	timeout := make(chan bool, 1)
	defer close(timeout)
	timer := time.NewTimer(30 * time.Second)

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
		case <-timer.C:
			if msg != nil {
				log.Println("Touching message")
				msg.Touch()
			}

		case <-sigChan:
			s.Stop()
			msg := fmt.Sprint("\nCanceled by user. count:", count)
			log.Fatal(msg)
		default:
		}
	}

	log.Println("Published", count, "points with a score >", s.MinScore*100, "% at", s.ScansPerSec, "scans/sec in", s.TotalTime)

	return nil
}
