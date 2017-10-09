package poller_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"net/url"

	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/poller"
	"github.com/stretchr/testify/assert"
)

func staticResponse(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, []byte(`{"test": 1}`))
}

func TestConnection_Connect(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(staticResponse))
	defer ts.Close()

	tsNonSsl := httptest.NewServer(http.HandlerFunc(staticResponse))
	defer tsNonSsl.Close()

	tests := []struct {
		name               string
		host               func() string
		tlsConfig          func(host string) *tls.Config
		guid               string
		expectedErr        bool
		expectedErrMessage string
		ctx                context.Context
	}{
		{
			name: "Happy path",
			host: func() string {
				testURL, _ := url.Parse(ts.URL)
				return testURL.Host
			},
			tlsConfig: func(host string) *tls.Config {
				return &tls.Config{
					InsecureSkipVerify: true,
					ServerName:         host,
					RootCAs:            nil,
				}
			},
			guid:        "happy-test",
			ctx:         context.Background(),
			expectedErr: false,
		},
		{
			name: "Invalid url",
			host: func() string {
				return "invalid-url"
			},
			guid:               "another-test",
			ctx:                context.Background(),
			expectedErr:        true,
			expectedErrMessage: "dial tcp: address invalid-url: missing port in address",
		},
		{
			name: "Empty context",
			host: func() string {
				testURL, _ := url.Parse(ts.URL)
				return testURL.Host
			},
			guid:               "empty-context-guid",
			ctx:                nil,
			expectedErr:        true,
			expectedErrMessage: "Context is undefined",
		},
		{
			name: "Cleartext connection",
			host: func() string {
				testURL, _ := url.Parse(tsNonSsl.URL)
				return testURL.Host
			},
			tlsConfig: func(host string) *tls.Config {
				return nil
			},
			guid:        "cleartext-guid",
			ctx:         context.Background(),
			expectedErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			reconciler := NewMockChecksReconciler(ctrl)

			conn := poller.NewConnection(tt.host(), tt.guid, reconciler)
			if tt.expectedErr {
				err := conn.Connect(tt.ctx, config.NewConfig("1-2-3", false, nil), nil)
				assert.Error(t, err)
			} else {
				assert.NoError(t,
					conn.Connect(tt.ctx,
						config.NewConfig("1-2-3", false, nil),
						tt.tlsConfig(tt.host())),
					"Simple connect should not throw an error")
			}
		})
	}
}
