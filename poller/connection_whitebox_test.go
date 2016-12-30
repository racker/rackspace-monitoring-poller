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
	var testStream = &ConnectionStream{}
	tests := []struct {
		name       string
		connection ConnectionInterface
		expected   ConnectionStreamInterface
	}{
		{
			name: "Happy path",
			connection: &Connection{
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
		stream            ConnectionStreamInterface
		session           *Session
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
				stream:            &ConnectionStream{},
				conn:              &net.TCPConn{},
				address:           "test",
				guid:              "test",
				connectionTimeout: 10 * time.Second,
			},
		},
		{
			name: "Wrong interface",
			fields: fields{
				stream:            &ConnectionStream{},
				conn:              &net.UDPConn{},
				address:           "test",
				guid:              "test",
				connectionTimeout: 10 * time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			conn := &Connection{
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
		stream            ConnectionStreamInterface
		session           *Session
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
				stream:            &ConnectionStream{},
				conn:              &net.TCPConn{},
				address:           "test",
				guid:              "test",
				connectionTimeout: 10 * time.Second,
			},
		},
		{
			name: "Wrong interface",
			fields: fields{
				stream:            &ConnectionStream{},
				conn:              &net.UDPConn{},
				address:           "test",
				guid:              "test",
				connectionTimeout: 10 * time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			conn := &Connection{
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
	mockSession := NewMockSessionInterface(mockCtrl)

	tests := []struct {
		name              string
		stream            *ConnectionStream
		session           *MockSessionInterface
		address           string
		guid              string
		connectionTimeout time.Duration
	}{
		{
			name:              "Happy path",
			stream:            &ConnectionStream{},
			session:           mockSession,
			address:           "test-addr",
			guid:              "test-guid",
			connectionTimeout: 1 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &Connection{
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
	mockSession := NewMockSessionInterface(mockCtrl)

	tests := []struct {
		name              string
		stream            *ConnectionStream
		session           *MockSessionInterface
		address           string
		guid              string
		connectionTimeout time.Duration
	}{
		{
			name:              "Happy path",
			stream:            &ConnectionStream{},
			session:           mockSession,
			address:           "test-addr",
			guid:              "test-guid",
			connectionTimeout: 1 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &Connection{
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
	conn := &Connection{
		address: "test",
		conn:    &net.TCPConn{},
		guid:    "test-guid",
	}

	got := conn.GetConnection()

	assert.Equal(t, conn, got, fmt.Sprintf("Expected %v but got %v", conn, got))
}
