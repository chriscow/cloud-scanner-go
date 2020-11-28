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

	"reticle/scanner"
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

	s := scanner.Session{}
	if err := json.Unmarshal(msg.Body, &s); err != nil {
		return err
	}

	scanner.RestoreSession(&s)

	log.Println("[Scan Radius Serivce] Received scan session request", s.ID, "for", s.ScansReq, "scans at", s.ZLine.Origin, "keeping the best", s.MinScore*100, "%")

	timeout := make(chan bool, 1)
	defer close(timeout)
	timer := time.NewTimer(30 * time.Second)

	running := true
	// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	go scanner.ScanAndPublish(h.ctx, &s, sigChan)

	for running {
		select {
		case <-timer.C:
			if msg != nil {
				log.Println("Touching message")
				msg.Touch()
			}
		case <-ctx.Done():
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

func main() error {
	checkEnv()

	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	handler := &scanRadiusHandler{}

	go scanner.startConsumer(ctx, "scan-session", "scan-radius-scanner", handler)

	select {
	case <-sigChan:
		cancel()
	default:
	}
}
