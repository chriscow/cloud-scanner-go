package main

import (
	"errors"
	"github.com/chriscow/cloud-scanner-go/scan"
)

func dbGetSession(id string) (*scan.Session, error) {
	return nil, errors.New("dbGetSession not implemented")
}

func dbNewSession(payload SessionPayload) {
}
