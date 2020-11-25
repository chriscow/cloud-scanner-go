package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/shamaton/msgpack"
	"github.com/urfave/cli/v2"
)

type scanRadiusHandler struct{}

func (h *scanRadiusHandler) HandleMessage(msg *nsq.Message) error {
	if len(msg.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		// In this case, a message with an empty body is simply ignored/discarded.
		return nil
	}

	s := Session{}
	if err := json.Unmarshal(msg.Body, &s); err != nil {
		return err
	}

	RestoreSession(&s)

	log.Println("[Scan Radius Serivce] Received scan session request", s.ID, "for", s.ScansReq, "scans at", s.ZLine.Origin, "keeping the best", s.MinScore*100, "%")

	// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	return startScan(&s, msg)
}

func scanRadiusSvc(ctx *cli.Context) error {
	handler := &scanRadiusHandler{}
	return startConsumer("scan-session", "scan-radius-scanner", handler)
}

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
