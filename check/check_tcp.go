package check

import (
	"errors"
	log "github.com/Sirupsen/logrus"
)

type TCPCheck struct {
	CheckBase
}

func NewTCPCheck(base *CheckBase) Check {
	check := &TCPCheck{CheckBase: *base}
	return check
}

func (ch *TCPCheck) Run() (*CheckResultSet, error) {
	log.Printf("Running TCP Check: %v\n", ch.GetId())
	return nil, errors.New("Not implimented")
}
