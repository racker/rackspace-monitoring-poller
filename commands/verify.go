package commands

import (
	"encoding/json"
	"fmt"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var (
	verifyRepeat int
	verifyPeriod int
)

var (
	VerifyCmd = &cobra.Command{
		Use:   "verify",
		Short: "Start the endpoint service",
		Run: func(cmd *cobra.Command, args []string) {
			checkFile, err := cmd.Flags().GetString("check")
			if err != nil {
				utils.Die(err, "Invalid 'check' value")
			}
			if checkFile != "" {
				ch, err := check.ReadCheckFromFile(checkFile)
				if err != nil {
					utils.Die(err, fmt.Sprintf("Invalid check definition in %s", checkFile))
				}

				for i := 0; i <= verifyRepeat; i++ {
					rs, err := ch.Run()
					if err != nil {
						utils.Die(err, "Failed to run check")
					}

					prettyJson := json.NewEncoder(os.Stdout)
					prettyJson.SetIndent("", "  ")
					err = prettyJson.Encode(rs)
					if err != nil {
						utils.Die(err, "Failed to format check")
					}

					if verifyRepeat > 0 {
						fmt.Println("Waiting until next execution...")
						time.Sleep(time.Second * time.Duration(verifyPeriod))
					}
				}

			}
		},
	}
)

func init() {
	VerifyCmd.Flags().String("check", "", "The location of a check definition file")
	VerifyCmd.Flags().IntVar(&verifyRepeat, "repeat", 0, "Number of times to repeat the check execution")
	VerifyCmd.Flags().IntVar(&verifyPeriod, "period", 10, "When repeating this value specifies the number of seconds between executions")
}
