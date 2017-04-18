package commands

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/spf13/cobra"
	"sync"
)

var (
	pingCmdConfig = struct {
		hosts   []string
		count   uint8
		rounds  int
		timeout uint64
	}{}
	PingCmd = &cobra.Command{
		Use:   "ping",
		Short: "Execute ping checks ondemand",
		Run: func(cmd *cobra.Command, args []string) {
			hosts, err := cmd.Flags().GetStringSlice("host")
			if err != nil {
				log.Fatal(err)
			}

			var wg sync.WaitGroup
			for i, host := range hosts {
				wg.Add(1)
				go pingCmdInstance(host, fmt.Sprintf("ch%05d", i), &wg)
			}
			wg.Wait()

		},
	}
)

func init() {
	PingCmd.Flags().StringSliceVar(&pingCmdConfig.hosts, "host", []string{}, "Host(s) to ping")
	PingCmd.Flags().Uint8Var(&pingCmdConfig.count, "count", 5, "The number of pings per check")
	PingCmd.Flags().IntVar(&pingCmdConfig.rounds, "rounds", 1, "The number of check rounds")
	PingCmd.Flags().Uint64Var(&pingCmdConfig.timeout, "timeout", 60, "The ping check timeout in seconds")
}

func pingCmdInstance(host string, checkId string, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < pingCmdConfig.rounds; i++ {
		var ch check.PingCheck
		ch.Details.Count = pingCmdConfig.count
		ch.Id = checkId
		ch.IpAddresses = map[string]string{
			"main": host,
		}
		var alias string = "main"
		ch.TargetAlias = &alias
		ch.TargetHostname = &host
		ch.Timeout = pingCmdConfig.timeout

		log.WithFields(log.Fields{
			"checkId": checkId,
			"host":    host,
		}).Info("Running")
		crs, err := ch.Run()
		if err != nil {
			log.Fatal("Ping check failed", checkId, err)
			return
		}

		log.WithFields(log.Fields{
			"checkId":   checkId,
			"resultSet": crs,
		}).Info("Ping check completed")
	}

}
