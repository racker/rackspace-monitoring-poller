//
// Copyright 2016 Rackspace
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS-IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//


package utils

import (
	"net"
	"time"
	"encoding/json"
)

const (
	// ReadKeepaliveTolerance is the division of ReadKeepalive that will be added when computing the read deadline.
	// It's a divisor because of durations being integer.
	ReadKeepaliveTolerance = 2
)

// SmartConn wraps a net.Conn with auto-logic for managing the read and write deadlines
//
// This implements the io.WriterCloser interface, so it can be used accordingly. It also partially implements
// the net.Conn interface.
type SmartConn struct {
	managedConn           net.Conn
	// Wraps ourselves in an encoder for use in WriteJSON
	encoder               *json.Encoder

	// Specifies how long each Write (or helper thereof) is allowed to execute before triggering the channel deadline
	WriteAllowance        time.Duration
	ReadKeepalive         time.Duration

	// An optional function that will get invoked according to HeartbeatSendInterval
	HeartbeatSender       func(sconn *SmartConn)
	HeartbeatSendInterval time.Duration
	heartbeatTimer        *time.Timer
}

// NewSmartConn creates a SmartConn that manages the connection c.
// The object can be further configured, but Start MUST be called before it is fully usable.
func NewSmartConn(c net.Conn) *SmartConn {
	return &SmartConn{managedConn: c}
}

// Start initiates the read keepalive/deadline management and heartbeat sending.
// Returns any error from the connection deadline setting.
func (c *SmartConn) Start() error {
	c.encoder = json.NewEncoder(c)

	if c.ReadKeepalive != 0 {
		err := c.observedRead()
		if err != nil {
			return err
		}
	}

	if c.HeartbeatSendInterval != 0 && c.HeartbeatSender != nil {
		c.heartbeatTimer = time.AfterFunc(c.HeartbeatSendInterval, func() {
			// TODO add error return to HeartbeatSender and gracefully degrade
			c.HeartbeatSender(c)
			c.heartbeatTimer.Reset(c.HeartbeatSendInterval)
		})
	}

	return nil
}

func (c *SmartConn) Close() error {
	if c.heartbeatTimer != nil {
		c.heartbeatTimer.Stop()
	}
	return c.managedConn.Close()
}

func (c *SmartConn) Write(p []byte) (n int, err error) {
	// pause heartbeat sender during this operation and then reset it after
	if c.heartbeatTimer != nil {
		if !c.heartbeatTimer.Stop() {
			<-c.heartbeatTimer.C
		}
		defer c.heartbeatTimer.Reset(c.HeartbeatSendInterval)
	}

	if c.WriteAllowance > 0 {
		c.managedConn.SetWriteDeadline(time.Now().Add(c.WriteAllowance))
		defer c.managedConn.SetWriteDeadline(time.Time{})
	}

	return c.managedConn.Write(p)
}

func (c *SmartConn) Read(b []byte) (n int, err error) {
	defer c.observedRead()
	return c.managedConn.Read(b)
}

func (c *SmartConn) RemoteAddr() net.Addr {
	return c.managedConn.RemoteAddr()
}

// WriteJSON is a convenience method to write line-oriented JSON to the connection.
func (c *SmartConn) WriteJSON(v interface{}) error {
	return c.encoder.Encode(v)
}

func (c *SmartConn) observedRead() error {

	if c.ReadKeepalive > 0 {
		// Allow some tolerance on the read deadline since a keepalive sent "on the dot" could technically fall
		// a little too close to a deadline set exactly at the same.
		return c.managedConn.SetReadDeadline(time.Now().Add(
			c.ReadKeepalive + (c.ReadKeepalive/ReadKeepaliveTolerance)))
	} else {
		return nil
	}

}
