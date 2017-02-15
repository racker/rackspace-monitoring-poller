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

	createCaPemFile(t)
	createIntegrationCertFile(t)
	createIntegrationKeyFile(t)

	defer deleteCaPemFile(t)
	defer deleteCertPemFile(t)
	defer deleteKeyPemFile(t)

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
				"testdata/endpoint-config.json", "--no-logfile"},
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
				"testdata/endpoint-config.noagents.json", "--no-logfile"},
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
				"testdata/endpoint-config.nocerts.json", "--no-logfile"},
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
				"testdata/endpoint-config.nobind.json", "--no-logfile"},
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
				os.Setenv(config.EnvDevCA, CaFileLocation)
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
}
