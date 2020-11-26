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

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	timeout := make(chan bool, 1)
	defer close(timeout)
	timer := time.NewTimer(30 * time.Second)

	running := true
	// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	go scanAndPublish(&s, sigChan)

	for running {
		select {

		case <-timer.C:
			if msg != nil {
				log.Println("Touching message")
				msg.Touch()
			}
		case <-sigChan:
			s.Stop()
			msg := fmt.Sprint("\nCanceled by user.")
			log.Fatal(msg)
		}
	}

	return nil
}

func scanRadiusSvc(ctx *cli.Context) error {
	handler := &scanRadiusHandler{}
	return startConsumer("scan-session", "scan-radius-scanner", handler)
}
