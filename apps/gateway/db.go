package main

import (
	"errors"
	"reticle/scan"
)

func dbGetSession(id string) (*scan.Session, error) {
	return nil, errors.New("dbGetSession not implemented")
}

func dbNewSession(payload SessionPayload) {
}
