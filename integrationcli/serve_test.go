package integrationcli

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"testing"

	"fmt"

	"os"

	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/assert"
)

func TestStartServe(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	localCert, err := ioutil.ReadFile("testdata/server-certs/cert.pem")
	if err != nil {
		t.Skip("Unable to read cert.pem from testdata/server-certs/cert.pem")
	}
	localKey, err := ioutil.ReadFile("testdata/server-certs/key.pem")
	if err != nil {
		t.Skip("Unable to read key.pem from testdata/server-certs/key.pem")
	}

	cert, _ := tls.X509KeyPair(localCert, localKey)
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	tlsListener, _ := tls.Listen("tcp", "127.0.0.1:0", tlsConfig)
	listenHost := tlsListener.Addr().(*net.TCPAddr).IP.String()
	listenPort := tlsListener.Addr().(*net.TCPAddr).Port

	fmt.Println(listenHost, listenPort)

	localEndpointCfg := []byte(
		fmt.Sprintf(`monitoring_token 0000000000000000000000000000000000000000000000000000000000000000.7777
monitoring_id agentA
monitoring_endpoints %s:%d
monitoring_private_zones pzA`, listenHost, listenPort))
	err = ioutil.WriteFile("testdata/local-endpoint.cfg", localEndpointCfg, 0644)
	if err != nil {
		t.Skip("Unable to write config file for happy path")
	}

	noAgentEndpointCfg := []byte(
		fmt.Sprintf(`monitoring_token 0000000000000000000000000000000000000000000000000000000000000000.7777
monitoring_endpoints %s:%d
monitoring_private_zones pzA`, listenHost, listenPort))
	err = ioutil.WriteFile("testdata/local-endpoint.noagent.cfg", noAgentEndpointCfg, 0644)
	if err != nil {
		t.Skip("Unable to write config file for no agent")
	}

	// Start TCP Server
	server := utils.NewBannerServer()
	go server.ServeTLS(tlsListener)

	tests := []struct {
		name           string
		args           []string
		expectedStdOut []*utils.OutputMessage
		expectedStdErr []*utils.OutputMessage
		runWithDevCa   bool
	}{
		{
			name: "Happy path",
			args: []string{
				"serve", "--config",
				"testdata/local-endpoint.cfg"},
			expectedStdOut: []*utils.OutputMessage{},
			expectedStdErr: []*utils.OutputMessage{
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Token",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Monitoring Id: agentA",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Zones: pzA",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Loaded configuration",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Assigned unique identifier",
				},
				&utils.OutputMessage{
					Level:   "info",
					Msg:     "Connecting to agent/poller endpoint",
					Address: fmt.Sprintf("%s:%d", listenHost, listenPort),
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "  ... Connected",
				},
			},
			runWithDevCa: true,
		},
		{
			name: "No token",
			args: []string{
				"serve", "--config",
				"testdata/local-endpoint.notoken.cfg"},
			expectedStdOut: []*utils.OutputMessage{},
			expectedStdErr: []*utils.OutputMessage{
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Monitoring Id: agentA",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Endpoints: somethingsomething:55000",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Zones: pzA",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Loaded configuration",
				},
				&utils.OutputMessage{
					Level: "error",
					Msg:   "Failed to validate configuration",
				},
				&utils.OutputMessage{
					Level: "error",
					Msg:   "Reason: No token is defined",
				},
			},
		},
		{
			name: "No agent",
			args: []string{
				"serve", "--config",
				"testdata/local-endpoint.noagent.cfg"},
			expectedStdOut: []*utils.OutputMessage{},
			expectedStdErr: []*utils.OutputMessage{
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Token",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   fmt.Sprintf("cfg: Setting Endpoints: %s:%d", listenHost, listenPort),
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Zones: pzA",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Loaded configuration",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Assigned unique identifier",
				},
				&utils.OutputMessage{
					Level:   "info",
					Msg:     "Connecting to agent/poller endpoint",
					Address: fmt.Sprintf("%s:%d", listenHost, listenPort),
				},
			},
		},
		{
			name: "No endpoints",
			args: []string{
				"serve", "--config",
				"testdata/local-endpoint.noendpoints.cfg"},
			expectedStdOut: []*utils.OutputMessage{},
			expectedStdErr: []*utils.OutputMessage{
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Token",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Monitoring Id: agentA",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Zones: pzA",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Loaded configuration",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Assigned unique identifier",
				},
				&utils.OutputMessage{
					Level:   "info",
					Msg:     "Connecting to agent/poller endpoint",
					Address: "agent-endpoint-ord.monitoring.api.rackspacecloud.com.:443",
				},
				&utils.OutputMessage{
					Level:   "info",
					Msg:     "Connecting to agent/poller endpoint",
					Address: "agent-endpoint-lon.monitoring.api.rackspacecloud.com.:443",
				},
				&utils.OutputMessage{
					Level:   "info",
					Msg:     "Connecting to agent/poller endpoint",
					Address: "agent-endpoint-dfw.monitoring.api.rackspacecloud.com.:443",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "  ... Connected",
				},
			},
		},
		{
			name: "No zones",
			args: []string{
				"serve", "--config",
				"testdata/local-endpoint.nozones.cfg"},
			expectedStdOut: []*utils.OutputMessage{},
			expectedStdErr: []*utils.OutputMessage{
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Token",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Monitoring Id: agentA",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Endpoints: somethingsomething:55000",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Loaded configuration",
				},
				&utils.OutputMessage{
					Level: "error",
					Msg:   "Failed to validate configuration",
				},
				&utils.OutputMessage{
					Level: "error",
					Msg:   "Reason: No zones are defined",
				},
			},
		},
		/*{
			name: "Endpoint not found",
			args: []string{
				"serve", "--config",
				"testdata/local-endpoint.endpointnotfound.cfg"},
			expectedStdOut: []*utils.OutputMessage{},
			expectedStdErr: []*utils.OutputMessage{
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Token",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Monitoring Id: agentA",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Endpoints: somethingsomething:5500",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "cfg: Setting Zones: pzA",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Loaded configuration",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Assigned unique identifier",
				},
				&utils.OutputMessage{
					Level:   "info",
					Msg:     "Connecting to agent/poller endpoint",
					Address: "somethingsomething:5500",
				},
				&utils.OutputMessage{
					Level: "error",
					Msg:   "Error: dial tcp: lookup somethingsomething: no such host",
				},
			},
		},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.runWithDevCa {
				os.Setenv(config.EnvDevCA, "testdata/server-certs/ca.pem")
			} else {
				os.Unsetenv(config.EnvDevCA)
			}

			result := runCmd(tt.args)
			gotOut := utils.BufferToStringSlice(result.StdOut)
			for entry := range tt.expectedStdOut {
				assert.Contains(t, gotOut, entry)
			}
			gotErr := utils.BufferToStringSlice(result.StdErr)
			for _, entry := range tt.expectedStdErr {
				assert.Contains(t, gotErr, entry)
			}
		})
	}

	server.Stop()
	tlsListener.Close()
}
