package main

import (
	"encoding/json"
	"log"
	"os"
	"github.com/chriscow/cloud-scanner-go/scan"
	"time"

	"github.com/go-chi/valve"
	"github.com/nsqio/go-nsq"
)

const channel = "qos"

type server struct {
	topic    string
	consumer *nsq.Consumer
	results  *scan.ScoredResults
	valve    *valve.Valve
}

func newServer(topic string) *server {
	v := valve.New()
	pub := make(chan []scan.Result)
	sr := scan.NewScoredResults(v.Context(), 100, pub)

	s := &server{
		topic:   topic,
		results: sr,
		valve:   v,
	}

	return s
}

func (s *server) start() error {
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(s.topic, channel, config)
	if err != nil {
		return err
	}

	consumer.AddHandler(s)

	// Use nsqlookupd to discover nsqd instances.
	// See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
	err = consumer.ConnectToNSQLookupd(os.Getenv("NSQ_LOOKUP") + ":4161")
	// err = consumer.ConnectToNSQD("localhost:4150")
	if err != nil {
		consumer.Stop()
		return err
	}

	s.consumer = consumer
	return nil
}

func (s *server) stop() {
	s.consumer.Stop()
	s.valve.Shutdown(5 * time.Second)
}

func (s *server) HandleMessage(msg *nsq.Message) error {
	if len(msg.Body) == 0 {
		return nil
	}

	var r scan.Result
	if err := json.Unmarshal(msg.Body, &r); err != nil {
		log.Println("JSON error:", err)
		return nil // don't requeue
	}

	s.results.Add(r)

	return nil
}
