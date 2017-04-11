package commands

import (
	"encoding/json"
	"fmt"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/spf13/cobra"
	"os"
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
			}
		},
	}
)

func init() {
	VerifyCmd.Flags().String("check", "", "The location of a check definition file")
}
