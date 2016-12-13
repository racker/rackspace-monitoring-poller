package utils_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/racker/rackspace-monitoring-poller/utils"
)

func TestDieWithMessages(t *testing.T) {
	// mock error output
	bak := utils.ErrOutput
	utils.ErrOutput = new(bytes.Buffer)
	defer func() { utils.ErrOutput = bak }()

	// mock exit code
	code := 0
	osexit := utils.Exit
	utils.Exit = func(c int) { code = c }
	defer func() { utils.Exit = osexit }()

	utils.Die(errors.New("failed"), "failure message")
	if utils.ErrOutput.(*bytes.Buffer).String() != "failure message\nReason: failed\n" {
		t.Fatal("Invalid error message")
	}

	if code != 1 {
		t.Fatal("Wrong exit code!")
	}
}

func TestDieWithMultipleMessages(t *testing.T) {
	// mock error output
	bak := utils.ErrOutput
	utils.ErrOutput = new(bytes.Buffer)
	defer func() { utils.ErrOutput = bak }()

	// mock exit code
	code := 0
	osexit := utils.Exit
	utils.Exit = func(c int) { code = c }
	defer func() { utils.Exit = osexit }()

	utils.Die(errors.New("failed"), "failure message one", "failure message two")
	if utils.ErrOutput.(*bytes.Buffer).String() != "failure message one\nfailure message two\nReason: failed\n" {
		t.Fatal("Invalid error message")
	}

	if code != 1 {
		t.Fatal("Wrong exit code!")
	}
}
