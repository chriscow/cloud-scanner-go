package main

import (
	"context"
	"encoding/json"
	"fmt"
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

type scanRadiusHandler struct {
	ctx context.Context
}

func NewHandler(ctx context.Context) *scanRadiusHandler {
	return &scanRadiusHandler{ctx: ctx}
}

func (h *scanRadiusHandler) HandleMessage(msg *nsq.Message) error {
	if len(msg.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		// In this case, a message with an empty body is simply ignored/discarded.
		return nil
	}

	s := scan.Session{}
	if err := json.Unmarshal(msg.Body, &s); err != nil {
		return err
	}

	scan.RestoreSession(&s)

	log.Println("[Scan Radius Serivce] Received scan session request", s.ID, "for", s.ScansReq, "scans at", s.ZLine.Origin, "keeping the best", s.MinScore*100, "%")

	timeout := make(chan bool, 1)
	defer close(timeout)
	timer := time.NewTimer(30 * time.Second)

	running := true
	// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	go scan.ScanAndPublish(h.ctx, &s)

	for running {
		select {
		case <-timer.C:
			if msg != nil {
				log.Println("Touching message")
				msg.Touch()
			}
		case <-h.ctx.Done():
			s.Stop()
			msg := fmt.Sprint("\nCanceled by user.")
			log.Fatal(msg)
		}
	}

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

	handler := &scanRadiusHandler{}

	log.Println("Watching for sessions on", sessionTopic, "publishing to", resultTopic)
	go util.StartConsumer(ctx, sessionTopic, resultTopic, handler)

	<-sigChan
	cancel()
	log.Println("\nUser cancelled")
}
