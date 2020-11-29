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
	sessionTopic   = "scan-session"        // subscribe to this topic for scan requests
	sessionChannel = "radius-scanner"      // channel for the above topic (who am I?)
	resultTopic    = "scan-radius-results" // publish results to this topic
	touchSec       = 30                    // touch the message every so often
)

type scanRadiusHandler struct{}

func (h scanRadiusHandler) HandleMessage(msg *nsq.Message) error {
	if len(msg.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		// In this case, a message with an empty body is simply ignored/discarded.
		return nil
	}

	log.Println("auto response:", msg.IsAutoResponseDisabled(), "has responded:", msg.HasResponded())

	s := scan.Session{}
	if err := json.Unmarshal(msg.Body, &s); err != nil {
		return err
	}

	if err := scan.Restore(&s); err != nil {
		return err
	}

	log.Println("[Scan Radius Serivce] Received scan session request", s.ID, "for", s.ScansReq, "scans at", s.ZLine.Origin, "keeping the best", s.MinScore*100, "%")

	cctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done, err := scan.Run(cctx, resultTopic, &s)
	if err != nil {
		log.Fatal(err)
	}

loop:
	for now := range time.Tick(touchSec * time.Second) {
		log.Println("Touching message", now)
		msg.Touch()
		select {
		case <-done:
			break loop
		default:
		}
	}

	return nil // auto-ack the msg
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
	go util.StartConsumer(ctx, sessionTopic, sessionChannel, handler)

	<-sigChan
	cancel()
	log.Println("\nUser cancelled")
}
