package main

import (
	"os"

	"github.com/nsqio/go-nsq"
)

const (
	wsChannel = "websocket"
)

type wsMsgHandler struct {
	hub *Hub
}

func (h *wsMsgHandler) HandleMessage(msg *nsq.Message) error {
	if len(msg.Body) == 0 {
		return nil
	}

	// broadcast the message to all connected clients
	h.hub.broadcast <- msg.Body

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
	h := &wsMsgHandler{}
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
