package main

import (
	"encoding/json"
	"log"

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

	// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	return startScan(&s, msg)
}

func scanRadiusSvc(ctx *cli.Context) error {
	handler := &scanRadiusHandler{}
	return startConsumer("scan-session", "scan-radius-scanner", handler)
}
