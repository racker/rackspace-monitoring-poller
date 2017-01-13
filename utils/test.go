package utils

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Timebox is used for putting a time bounds around a chunk of code, given as the function boxed.
// NOTE that if the duration d elapses, then boxed will be left to run off in its go-routine...it can't be
// forcefully terminated.
// This function can be used outside of a unit test context by passing nil for t
// Returns true if boxed finished before duration d elapsed.
func Timebox(t *testing.T, d time.Duration, boxed func(t *testing.T)) bool {
	timer := time.NewTimer(d)
	completed := make(chan struct{})

	go func() {
		boxed(t)
		close(completed)
	}()

	select {
	case <-timer.C:
		if t != nil {
			t.Fatal("Timebox expired")
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
	log.Debug(conn.RemoteAddr(), "connected")
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
			log.Debug("disconnecting", conn.RemoteAddr())
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
