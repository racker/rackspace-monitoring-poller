package integrationcli

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/assert"
)

func TestStartServeEndpointHappyPath(t *testing.T) {
	t.Skip("Currently this is not working due to endpoint protocol being in flux.")
	if testing.Short() {
		t.Skip()
	}
	// start endpoint - validate it's up and listening to requests
	endpointTimeout := time.Duration(120 * time.Second)
	endpointDone := make(chan *utils.Result, 1)
	go runEndpoint(endpointTimeout, endpointDone)

	// start serve - validate it's up and able to communicate with endpoint
	serveTimeout := time.Duration(120 * time.Second)
	serveDone := make(chan *utils.Result, 1)
	os.Setenv(config.EnvDevCA, "testdata/server-certs/ca.pem")
	go runServe(serveTimeout, serveDone)

	endpointResult := <-endpointDone
	// validate the endpoint is listening
	fmt.Println("endpoint - we done")
	endpointStdErr := utils.BufferToStringSlice(endpointResult.StdErr)
	assert.Contains(t, endpointStdErr, &utils.OutputMessage{
		Level: "info",
		Msg:   "Endpoint is accepting connections from pollers",
		BoundAddr: &utils.BoundAddress{
			IP:   "::",
			Port: 55000,
			Zone: "",
		},
	})

	serveResult := <-serveDone
	// validate the server
	fmt.Println("serve - we done")
	serveStdErr := utils.BufferToStringSlice(serveResult.StdErr)
	assert.Contains(t, serveStdErr, &utils.OutputMessage{
		Level: "info",
		Msg:   "  ... Connected",
	})

	// wait and listen for metrics on endpoint and serve
}

func runEndpoint(endpointTimeout time.Duration, endpointDone chan *utils.Result) {
	endpointDone <- runCmdWithTimeout([]string{
		"endpoint", "--config",
		"testdata/endpoint-config.happypath.json"},
		false, endpointTimeout)
}

func runServe(serveTimeout time.Duration, serveDone chan *utils.Result) {
	serveDone <- runCmdWithTimeout([]string{
		"serve", "--config",
		"testdata/local-endpoint.happypath.cfg"},
		false, serveTimeout)

}
