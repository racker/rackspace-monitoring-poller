package utils

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"

	"bufio"

	"encoding/json"

	"errors"

	"os"

	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const BuildDir = "build"

// PollerCommand is the program name for the poller.
// Used for integration tests
var PollerCommand = fmt.Sprintf("%s/src/github.com/racker/rackspace-monitoring-poller/%s/rackspace-monitoring-poller_%s_%s",
	os.Getenv("GOPATH"),
	BuildDir,
	runtime.GOOS,
	runtime.GOARCH,
)

// Commander interface is a wrapper around *exec.Cmd.
// Mainly utilized for testing subprocesses without executing code
type Commander interface {
	Start() error
	Wait() error
	GetProcess() *os.Process
	GetProcessState() *os.ProcessState
	GetArgs() []string
}

type PollerCmd struct {
	*exec.Cmd
}

func (c *PollerCmd) GetProcess() *os.Process {
	return c.Process
}

func (c *PollerCmd) GetProcessState() *os.ProcessState {
	return c.ProcessState
}

func (c *PollerCmd) GetArgs() []string {
	return c.Args
}

// Result is used to wrap integration test results
// It wraps the command struct that sets up the required timeout
// It also wraps STDOUT and STDERR for output validation,
// exit codes for validation, and timeout that will either
// kill your process and report error or kill your process and
// allow you to validate outputs
type Result struct {
	Command        Commander
	Error          error
	ErrorOnTimeout bool
	Timeout        time.Duration
	StdOut         *bytes.Buffer
	StdErr         *bytes.Buffer
}

type OutputMessage struct {
	Level     string
	Msg       string
	Address   string
	BoundAddr *BoundAddress
}

type BoundAddress struct {
	IP   string
	Port int
	Zone string
}

// SetupCommand sets up Result with passed in command list,
// and timeouts.
// timeout for erroring takes preference over timeout success
func SetupCommand(commandList []string,
	errorOnTimeout bool, timeout time.Duration) *Result {
	log.Println("Setup command")
	log.Println(commandList)
	internalCmd := exec.Command(PollerCommand, commandList...)
	cmd := &PollerCmd{internalCmd}
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	return &Result{
		Command:        cmd,
		StdOut:         &outbuf,
		StdErr:         &errbuf,
		ErrorOnTimeout: errorOnTimeout,
		Timeout:        timeout,
	}
}

// StartCommand attempts to start the command.
// Sets the result error if start fails
func StartCommand(result *Result) {
	log.Println("Start command")
	err := result.Command.Start()
	if err != nil {
		result.Error = err
	}
}

func RunCommand(result *Result) {
	log.Println("Run command")
	done := make(chan error, 1)
	// Wait for command to exit in a goroutine
	go func() {
		done <- result.Command.Wait()
	}()

	select {
	case <-time.After(result.Timeout):
		killErr := result.Command.GetProcess().Kill()
		if killErr != nil {
			fmt.Printf("failed to kill (pid=%d): %v\n", result.Command.GetProcess().Pid, killErr)
			result.Error = killErr
		}
		if result.ErrorOnTimeout {
			// we should not have timed out.  Oops!
			result.Error = errors.New("Failed on timeout!")
		}
		log.Printf("success state: %v", result.Command.GetProcessState().String())
		log.Printf("Stdout: %v", result.StdOut.String())
		log.Printf("Stderr: %v", result.StdErr.String())

	case err := <-done:
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				log.Printf("success state: %v", result.Command.GetProcessState().String())
				log.Printf("Stdout: %v", result.StdOut.String())
				log.Printf("Stderr: %v", result.StdErr.String())
				log.Printf("Exit Status: %d", status.ExitStatus())
			}
		} else {
			result.Error = err
			log.Errorf("We finished but threw an exit error: %v", err)
		}
	}
}

func BufferToStringSlice(buf *bytes.Buffer) []*OutputMessage {
	var messageList = []*OutputMessage{}
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		var outputMessage = &OutputMessage{}
		err := json.Unmarshal(scanner.Bytes(), outputMessage)

		if err != nil {
			log.Errorf("Unable to parse JSON: %v", err)
		} else {
			messageList = append(messageList, outputMessage)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return messageList
}

// Timebox is used for putting a time bounds around a chunk of code, given as the function boxed.
// NOTE that if the duration d elapses, then boxed will be left to run off in its go-routine...it can't be
// forcefully terminated.
// This function can be used outside of a unit test context by passing nil for t
// Returns true if boxed finished before duration d elapsed.
func Timebox(t *testing.T, d time.Duration, boxed func(t *testing.T)) bool {
	return TimeboxNamed(t, "", d, boxed)
}

// Timebox is used for putting a time bounds around a chunk of code, given as the function boxed.
// NOTE that if the duration d elapses, then boxed will be left to run off in its go-routine...it can't be
// forcefully terminated.
// This function can be used outside of a unit test context by passing nil for t
// Returns true if boxed finished before duration d elapsed.
func TimeboxNamed(t *testing.T, name string, d time.Duration, boxed func(t *testing.T)) bool {
	timer := time.NewTimer(d)
	completed := make(chan struct{})

	go func() {
		boxed(t)
		close(completed)
	}()

	select {
	case <-timer.C:
		if t != nil {
			t.Fatal("Timebox expired", name)
		}
		return false
	case <-completed:
		timer.Stop()
		return true
	}
}

func TestTimebox_Quick(t *testing.T) {
	result := Timebox(t, 1*time.Second, func(t *testing.T) {
		time.Sleep(1 * time.Millisecond)
	})
	assert.True(t, result)
}

func TestTimebox_TimesOut(t *testing.T) {
	result := Timebox(nil, 1*time.Millisecond, func(t *testing.T) {
		time.Sleep(100 * time.Millisecond)
	})

	assert.False(t, result)
}

type BannerServer struct {
	HandleConnection func(conn net.Conn)

	waitGroup *sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewBannerServer() *BannerServer {
	server := &BannerServer{}
	server.waitGroup = &sync.WaitGroup{}
	server.ctx, server.cancel = context.WithCancel(context.Background())
	server.HandleConnection = server.serve
	return server
}

func (s *BannerServer) Stop() {
	s.cancel()
	s.waitGroup.Wait()
}

func (s *BannerServer) Serve(listener net.Listener) {
	conn, err := listener.Accept()
	if nil != err {
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			return
		}
		log.WithField("err", err).Fatal("Unexpected error")
	}
	log.WithField("remoteAddr", conn.RemoteAddr()).Debug("accepted")
	s.waitGroup.Add(1)
	go s.serve(conn)
}

func (s *BannerServer) ServeTLS(listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			return
		}
		log.WithField("err", err).Fatal("Unexpected error")
	}
	s.waitGroup.Add(1)
	go s.serve(conn)
}

func (s *BannerServer) serve(conn net.Conn) {
	defer s.waitGroup.Done()
	defer conn.Close()
	for {
		select {
		case <-s.ctx.Done():
			log.Println("disconnecting", conn.RemoteAddr())
			return
		default:
		}
		buf := make([]byte, 4096)
		conn.SetDeadline(time.Now().Add(1e9))
		conn.Write([]byte("SSH-2.0-OpenSSH_7.3\n"))
		conn.SetDeadline(time.Now().Add(1 * time.Second))
		if _, err := conn.Read(buf); nil != err {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			return
		}
		return
	}
}
