package main

import (
	"fmt"

	"github.com/nsqio/go-nsq"
	"github.com/shamaton/msgpack"
)

// Result holds the data from a single scan and is serialized with MessagePack
type Result struct {
	SessionID     int64
	Origin        Vector2
	ZeroType      ZeroType
	ZerosCount    int
	ZerosHit      int
	BestTheta     float64
	BestBucket    int
	AvgParity     float64
	LatticeParams interface{}
	Score         float64
}

func (r Result) String() string {
	return fmt.Sprint("session id:", r.SessionID, " origin:", r.Origin, r.ZeroType,
		" zeros:", r.ZerosCount, " hits:", r.ZerosHit, " theta:", r.BestTheta,
		" bucket:", r.BestBucket)
}

type scanResultHandler struct{}

// HandleMessage implements the Handler interface.
func (h scanResultHandler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		// In this case, a message with an empty body is simply ignored/discarded.
		return nil
	}

	// do whatever actual message processing is desired
	res := Result{}
	if err := msgpack.Decode(m.Body, &res); err != nil {
		fmt.Println("Failed to decode message", err)
		return err
	}

	fmt.Println(res.Origin)

	// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	return nil
}
