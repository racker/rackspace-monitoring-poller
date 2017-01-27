package poller

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestConnection_GetStream(t *testing.T) {
	var testStream = &EleConnectionStream{}
	tests := []struct {
		name       string
		connection Connection
		expected   ConnectionStream
	}{
		{
			name: "Happy path",
			connection: &EleConnection{
				stream: testStream,
			},
			expected: testStream,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := tt.connection
			got := conn.GetStream()
			assert.Equal(t, tt.expected, got, fmt.Sprintf("Connection.GetStream() = %v, expected %v", got, tt.expected))
		})
	}
}

func TestConnection_SetReadDeadline(t *testing.T) {
	type fields struct {
		stream            ConnectionStream
		session           Session
		conn              io.ReadWriteCloser
		address           string
		guid              string
		connectionTimeout time.Duration
	}
	tests := []struct {
		name     string
		fields   fields
		deadline time.Time
	}{
		{
			name: "Happy path",
			fields: fields{
				stream:            &EleConnectionStream{},
				conn:              &net.TCPConn{},
				address:           "test",
				guid:              "test",
				connectionTimeout: 10 * time.Second,
			},
		},
		{
			name: "Wrong interface",
			fields: fields{
				stream:            &EleConnectionStream{},
				conn:              &net.UDPConn{},
				address:           "test",
				guid:              "test",
				connectionTimeout: 10 * time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			conn := &EleConnection{
				address:           tt.fields.address,
				conn:              tt.fields.conn,
				stream:            tt.fields.stream,
				guid:              tt.fields.guid,
				connectionTimeout: tt.fields.connectionTimeout,
			}
			assert.NotPanics(t, func() { conn.SetReadDeadline(tt.deadline) }, "Calling conn.SetReadDeadline should not panic")

		})
	}
}

func TestConnection_SetWriteDeadline(t *testing.T) {
	type fields struct {
		stream            ConnectionStream
		session           Session
		conn              io.ReadWriteCloser
		address           string
		guid              string
		connectionTimeout time.Duration
	}
	tests := []struct {
		name     string
		fields   fields
		deadline time.Time
	}{
		{
			name: "Happy path",
			fields: fields{
				stream:            &EleConnectionStream{},
				conn:              &net.TCPConn{},
				address:           "test",
				guid:              "test",
				connectionTimeout: 10 * time.Second,
			},
		},
		{
			name: "Wrong interface",
			fields: fields{
				stream:            &EleConnectionStream{},
				conn:              &net.UDPConn{},
				address:           "test",
				guid:              "test",
				connectionTimeout: 10 * time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			conn := &EleConnection{
				address:           tt.fields.address,
				conn:              tt.fields.conn,
				stream:            tt.fields.stream,
				guid:              tt.fields.guid,
				connectionTimeout: tt.fields.connectionTimeout,
			}
			assert.NotPanics(t, func() { conn.SetWriteDeadline(tt.deadline) }, "Calling conn.SetWriteDeadline should not panic")

		})
	}
}

func TestConnection_Close(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockSession := NewMockSession(mockCtrl)

	tests := []struct {
		name              string
		stream            *EleConnectionStream
		session           *MockSession
		address           string
		guid              string
		connectionTimeout time.Duration
	}{
		{
			name:              "Happy path",
			stream:            &EleConnectionStream{},
			session:           mockSession,
			address:           "test-addr",
			guid:              "test-guid",
			connectionTimeout: 1 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &EleConnection{
				stream:            tt.stream,
				session:           tt.session,
				address:           tt.address,
				guid:              tt.guid,
				connectionTimeout: tt.connectionTimeout,
			}
			mockSession.EXPECT().Close()
			conn.Close()
		})
	}
}

func TestConnection_Wait(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockSession := NewMockSession(mockCtrl)

	tests := []struct {
		name              string
		stream            *EleConnectionStream
		session           *MockSession
		address           string
		guid              string
		connectionTimeout time.Duration
	}{
		{
			name:              "Happy path",
			stream:            &EleConnectionStream{},
			session:           mockSession,
			address:           "test-addr",
			guid:              "test-guid",
			connectionTimeout: 1 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &EleConnection{
				stream:            tt.stream,
				session:           tt.session,
				address:           tt.address,
				guid:              tt.guid,
				connectionTimeout: tt.connectionTimeout,
			}
			mockSession.EXPECT().Wait()
			conn.Wait()
		})
	}
}

func TestGetConnection(t *testing.T) {
	tcpConn := &net.TCPConn{}

	conn := &EleConnection{
		address: "test",
		conn:    tcpConn,
		guid:    "test-guid",
	}

	writer := conn.GetFarendWriter()
	assert.Equal(t, tcpConn, writer, fmt.Sprintf("Expected %v but got %v", tcpConn, writer))

	reader := conn.GetFarendReader()
	assert.Equal(t, tcpConn, reader, fmt.Sprintf("Expected %v but got %v", tcpConn, writer))
}
