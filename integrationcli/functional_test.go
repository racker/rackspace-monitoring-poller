package integrationcli

import (
	"fmt"
	"log"
	"time"

	"github.com/racker/rackspace-monitoring-poller/utils"
)

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
