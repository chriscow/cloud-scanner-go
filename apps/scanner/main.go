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
	scannerChannel = "scanner" // channel for the above topic (who am I?)
	touchSec       = 30        // touch the message every so often
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

	done, err := scan.Run(cctx, scan.ResultTopic, &s)
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

	sessionComplete(s)
	return nil // auto-ack the msg
}

func sessionComplete(s scan.Session) error {
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		return err
	}
	defer producer.Stop()

	body, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return producer.Publish(scan.CompleteTopic, body)
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

	log.Println("Watching for sessions on", scan.SessionTopic, "publishing to", scan.ResultTopic)
	go util.StartConsumer(ctx, scan.SessionTopic, scannerChannel, handler)

	<-sigChan
	cancel()
	log.Println("\nUser cancelled")
}
