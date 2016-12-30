package poller_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"net/url"

	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/poller"
	"github.com/stretchr/testify/assert"
)

func staticResponse(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, []byte(`{"test": 1}`))
}

func TestNewConnection(t *testing.T) {
	type args struct {
		address string
		guid    string
		stream  poller.ConnectionStreamInterface
	}
	var testStream = &poller.ConnectionStream{}
	tests := []struct {
		name     string
		args     args
		expected *poller.ConnectionStream
	}{
		{
			name: "Happy path",
			args: args{
				address: "http://example.com",
				guid:    "my-pure-awesomeness",
				stream:  testStream,
			},
			expected: testStream,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := poller.NewConnection(tt.args.address, tt.args.guid, tt.args.stream)
			//validate stream
			assert.Equal(t, tt.expected, got.GetStream(), fmt.Sprintf("NewConnection() stream = %v, expected %v", got, tt.expected))
		})
	}
}

func TestConnection_Connect(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(staticResponse))
	defer ts.Close()

	tests := []struct {
		name               string
		url                func() string
		guid               string
		stream             poller.ConnectionStreamInterface
		expectedErr        bool
		expectedErrMessage string
		ctx                context.Context
	}{
		{
			name: "Happy path",
			url: func() string {
				testURL, _ := url.Parse(ts.URL)
				return testURL.Host
			},
			guid: "happy-test",
			stream: poller.NewConnectionStream(&config.Config{
				Guid: "test-guid",
			}),
			ctx:         context.Background(),
			expectedErr: false,
		},
		{
			name: "Invalid url",
			url: func() string {
				return "invalid-url"
			},
			guid: "another-test",
			stream: poller.NewConnectionStream(&config.Config{
				Guid: "test-guid",
			}),
			ctx:                context.Background(),
			expectedErr:        true,
			expectedErrMessage: "dial tcp: missing port in address invalid-url",
		},
		{
			name: "Empty context",
			url: func() string {
				testURL, _ := url.Parse(ts.URL)
				return testURL.Host
			},
			guid: "empty-context-guid",
			stream: poller.NewConnectionStream(&config.Config{
				Guid: "test-guid",
			}),
			ctx:                nil,
			expectedErr:        true,
			expectedErrMessage: "Context is undefined",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := poller.NewConnection(tt.url(), tt.guid, tt.stream)
			if tt.expectedErr {
				err := conn.Connect(tt.ctx)
				assert.EqualError(
					t, err, tt.expectedErrMessage,
					fmt.Sprintf("Expected to throw %v but got %v", tt.expectedErrMessage, err))
			} else {
				assert.NoError(t, conn.Connect(tt.ctx), "Simple connect should not throw an error")
			}
		})
	}
}
