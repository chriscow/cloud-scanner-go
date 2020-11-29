package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/nsqio/go-nsq"

	"reticle/scan"
	"reticle/util"
)

const (
	sessionTopic = "scan-session"
	resultTopic  = "scan-radius-results"
)

type scanRadiusHandler struct{}

func (h scanRadiusHandler) HandleMessage(msg *nsq.Message) error {
	if len(msg.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		// In this case, a message with an empty body is simply ignored/discarded.
		return nil
	}

	s := scan.Session{}
	if err := json.Unmarshal(msg.Body, &s); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	scan.Restore(ctx, &s)

	log.Println("[Scan Radius Serivce] Received scan session request", s.ID, "for", s.ScansReq, "scans at", s.ZLine.Origin, "keeping the best", s.MinScore*100, "%")

	timeout := make(chan bool, 1)
	defer close(timeout)
	timer := time.NewTimer(30 * time.Second)

	running := true

	go scan.Publish(ctx, resultTopic, &s)

loop:
	for running {
		select {
		case <-timer.C:
			if msg != nil {
				log.Println("Touching message")
				msg.Touch()
			}
		case <-ctx.Done():
			log.Println("\nCanceled by user.")
			break loop
		}
	}

	cancel()
	return nil
}

func checkEnv() {
	godotenv.Load()

	if os.Getenv("NSQ_LOOKUP") == "" {
		log.Fatal("NSQ_LOOKUP environment variable not set")
	}
}

func main() {
	checkEnv()

	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	handler := scanRadiusHandler{}

	log.Println("Watching for sessions on", sessionTopic, "publishing to", resultTopic)
	go util.StartConsumer(ctx, sessionTopic, resultTopic, handler)

	<-sigChan
	cancel()
	log.Println("\nUser cancelled")
}
