// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/go-chi/valve"
	"log"
	"os"
	"github.com/nsqio/go-nsq"
)


const wsChannel = "websocket"


type publication struct {
	valve  *valve.Valve
	topic string

	// registered subscribers.
	subscribers map[*subscriber]bool

	// broadcast messages to the subscribers.
	broadcast chan []byte

	// subscription requests from subscribers.
	subscribe chan *subscriber

	// unsubscribe requests from subscribers.
	unsubscribe chan *subscriber

	// nsq message consumer
	consumer   *nsq.Consumer
}

func newPublication(v *valve.Valve, topic string) *publication {
	return &publication{
		valve: v,
		topic: topic,
		broadcast:  make(chan []byte),
		subscribe:   make(chan *subscriber),
		unsubscribe: make(chan *subscriber),
		subscribers:    make(map[*subscriber]bool),
	}
}

func (p *publication) remove(s *subscriber) {
	delete(p.subscribers, s) // protected by channels
	s.cancel()

	if len(p.subscribers) == 0 && p.consumer != nil {
		p.consumer.Stop()
		p.consumer = nil
	}
}

func (p *publication) run() {
	p.valve.Open()
	defer p.valve.Close()

	for {
		select {
		case <-p.valve.Stop():
			for sub := range p.subscribers {
				p.remove(sub)
			}
			return
		case sub := <-p.subscribe:
			p.subscribers[sub] = true
			var err error
			if p.consumer == nil {
				err = p.startConsumer()
				if err != nil {
					log.Fatal("[publication] failed to start consumer", err)
				}
			}
		case sub := <-p.unsubscribe:
			if _, ok := p.subscribers[sub]; ok {
				p.remove(sub)
			}
		case message := <-p.broadcast:
			for sub := range p.subscribers {
				select {
				case sub.send <- message:
				default:
					p.remove(sub)
				}
			}
		}
	}
}


func (p *publication) HandleMessage(msg *nsq.Message) error {
	if len(msg.Body) == 0 {
		return nil
	}

	// broadcast the message to all connected subscribers
	if len(p.subscribers) > 0 {
		p.broadcast <- msg.Body
	}

	return nil
}

func (p *publication) startConsumer() (err error) {
	config := nsq.NewConfig()
	p.consumer, err = nsq.NewConsumer(p.topic, wsChannel, config)
	if err != nil {
		return err
	}

	p.consumer.AddHandler(p)

	// Use nsqlookupd to discover nsqd instances.
	// See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
	err = p.consumer.ConnectToNSQLookupd(os.Getenv("NSQ_LOOKUP") + ":4161")
	// err = consumer.ConnectToNSQD("localhost:4150")
	if err != nil {
		p.consumer.Stop()
		return err
	}

	return nil
}
