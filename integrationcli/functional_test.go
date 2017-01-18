package integrationcli

import (
	"fmt"
	"log"
	"testing"
	"time"

	"io/ioutil"
	"os"

	"github.com/racker/rackspace-monitoring-poller/utils"
)

var CaFileLocation = "testdata/ca.pem"
var CertFileLocation = "testdata/cert.pem"
var KeyFileLocation = "testdata/key.pem"

func createCaPemFile(t *testing.T) {
	createFile(t, CaFileLocation, utils.IntegrationTestCA)
}

func createIntegrationCertFile(t *testing.T) {
	createFile(t, CertFileLocation, utils.IntegrationTestCert)
}

func createIntegrationKeyFile(t *testing.T) {
	createFile(t, KeyFileLocation, utils.IntegrationTestKey)
}

func createFile(t *testing.T, location string, content []byte) {
	err := ioutil.WriteFile(location, content, 0644)
	if err != nil {
		t.Fatalf("Unable to create %s", location)
	}
}

func deleteCaPemFile(t *testing.T) {
	deleteFile(t, CaFileLocation)
}

func deleteKeyPemFile(t *testing.T) {
	deleteFile(t, KeyFileLocation)
}

func deleteCertPemFile(t *testing.T) {
	deleteFile(t, CertFileLocation)
}

func deleteFile(t *testing.T, location string) {
	err := os.Remove(location)
	if err != nil {
		t.Fatalf("Unable to remove %s", location)
	}
}

func runCmdWithTimeout(args []string, errorOnTimeout bool, timeout time.Duration) *utils.Result {
	result := utils.SetupCommand(
		args,
		errorOnTimeout,
		timeout,
	)

	utils.StartCommand(result)

	fmt.Printf("result %v", result)

	if result.Error != nil {
		log.Fatalf("cmd.Start: %v", result.Error)
	}

	utils.RunCommand(result)

	if result.Error != nil {
		log.Fatalf("cmd.Run: %v", result.Error)
	}

	return result

}

func runCmd(args []string) *utils.Result {
	log.Println("Run command")

	return runCmdWithTimeout(args, false, time.Duration(1*time.Second))

}
