package main

import (
	"os"
	"os/signal"
	"github.com/chriscow/cloud-scanner-go/scan"
	"syscall"
)

func main() {
	run(os.Args)
}

func run(args []string) error {
	s := newServer(scan.ResultTopic)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	s.start()
	<-sigChan
	s.stop()

	return nil
}
