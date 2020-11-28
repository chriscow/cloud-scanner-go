package scanner

import (
	"context"
	"os"

	"github.com/nsqio/go-nsq"
)

func StartConsumer(ctx context.Context, topic, channel string, handler nsq.Handler) error {
	// Instantiate a consumer that will subscribe to the provided channel.
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		return err
	}

	// Set the Handler for messages received by this Consumer. Can be called multiple times.
	// See also AddConcurrentHandlers.
	consumer.AddHandler(handler)

	// Use nsqlookupd to discover nsqd instances.
	// See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
	err = consumer.ConnectToNSQLookupd(os.Getenv("NSQ_LOOKUP") + ":4161")
	// err = consumer.ConnectToNSQD("localhost:4150")
	if err != nil {
		return err
	}

	// wait for signal to exit
	<-ctx.Done()

	// Gracefully stop the consumer.
	consumer.Stop()
	return nil
}
