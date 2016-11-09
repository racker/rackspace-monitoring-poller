package main

import (
	"errors"
	log "github.com/Sirupsen/logrus"
)

type PingCheck struct {
	CheckBase
}

func NewPingCheck(base *CheckBase) Check {
	check := &PingCheck{CheckBase: *base}
	check.PrintDefaults()
	return check
}

func (ch *PingCheck) Run() (*CheckResultSet, error) {
	log.Printf("Running PING Check: %v", ch.GetId())
	return nil, errors.New("Not implemented")
}
