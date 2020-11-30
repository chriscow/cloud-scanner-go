package main

import (
	"encoding/json"
	"log"
	"os"
	"reticle/scan"

	"github.com/nsqio/go-nsq"
)

const (
	wsChannel = "websocket"
)

type wsMsgHandler struct {
	hub     *Hub
	topic   string
	channel string
}

func (h *wsMsgHandler) HandleMessage(msg *nsq.Message) error {
	if len(msg.Body) == 0 {
		return nil
	}

	r := scan.Result{}
	err := json.Unmarshal(msg.Body, &r)
	if err != nil {
		log.Println(err)
	}

	// broadcast the message to all connected clients
	if len(h.hub.clients) > 0 {
		h.hub.broadcast <- msg.Body
	}

	return nil
}

func startConsumer(hub *Hub, topic, channel string) (*nsq.Consumer, error) {
	// Instantiate a consumer that will subscribe to the provided channel.
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		return nil, err
	}

	// Set the Handler for messages received by this Consumer. Can be called multiple times.
	// See also AddConcurrentHandlers.
	h := &wsMsgHandler{hub: hub, topic: topic, channel: channel}
	consumer.AddHandler(h)

	// Use nsqlookupd to discover nsqd instances.
	// See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
	err = consumer.ConnectToNSQLookupd(os.Getenv("NSQ_LOOKUP") + ":4161")
	// err = consumer.ConnectToNSQD("localhost:4150")
	if err != nil {
		consumer.Stop()
		return nil, err
	}

	return consumer, nil
}
