package integrationcli

import (
	"testing"

	"os"

	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/assert"
)

func TestStartEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	/*
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

			// Start TCP Server
			server := utils.NewBannerServer()
			go server.ServeTLS(tlsListener)
	*/
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
				"endpoint", "--config",
				"testdata/endpoint-config.json"},
			expectedStdOut: []*utils.OutputMessage{},
			expectedStdErr: []*utils.OutputMessage{
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Loaded configuration",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Starting metrics decomposer",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Starting agent tracker",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Endpoint is accepting connections from pollers",
					BoundAddr: &utils.BoundAddress{
						IP:   "::",
						Port: 55000,
						Zone: "",
					},
				},
			},
			runWithDevCa: true,
		},
		{
			name: "No agent configured",
			args: []string{
				"endpoint", "--config",
				"testdata/endpoint-config.noagents.json"},
			expectedStdOut: []*utils.OutputMessage{},
			expectedStdErr: []*utils.OutputMessage{
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Loaded configuration",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Starting metrics decomposer",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Starting agent tracker",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Endpoint is accepting connections from pollers",
					BoundAddr: &utils.BoundAddress{
						IP:   "::",
						Port: 55000,
						Zone: "",
					},
				},
			},
			runWithDevCa: true,
		},
		{
			name: "No certs specified",
			args: []string{
				"endpoint", "--config",
				"testdata/endpoint-config.nocerts.json"},
			expectedStdOut: []*utils.OutputMessage{},
			expectedStdErr: []*utils.OutputMessage{
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Loaded configuration",
				},
				&utils.OutputMessage{
					Level: "error",
					Msg:   "Invalid endpoint setup",
				},
				&utils.OutputMessage{
					Level: "error",
					Msg:   "Reason: Bad configuration: Missing CertFile",
				},
			},
			runWithDevCa: true,
		},
		{
			name: "No bind port specified (uses default port)",
			args: []string{
				"endpoint", "--config",
				"testdata/endpoint-config.nobind.json"},
			expectedStdOut: []*utils.OutputMessage{},
			expectedStdErr: []*utils.OutputMessage{
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Loaded configuration",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Starting metrics decomposer",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Starting agent tracker",
				},
				&utils.OutputMessage{
					Level: "info",
					Msg:   "Endpoint is accepting connections from pollers",
					BoundAddr: &utils.BoundAddress{
						IP:   "::",
						Port: 50041,
						Zone: "",
					},
				},
			},
			runWithDevCa: true,
		},
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

	//server.Stop()
	//tlsListener.Close()
}
