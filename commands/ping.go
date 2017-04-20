package commands

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/check"
	protocheck "github.com/racker/rackspace-monitoring-poller/protocol/check"
	"github.com/spf13/cobra"
	"strings"
	"sync"
)

var (
	pingCmdConfig = struct {
		hosts   []string
		count   uint8
		rounds  int
		timeout uint64
		ipType  string
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
	PingCmd.Flags().StringVar(&pingCmdConfig.ipType, "ip-type", "", "The IP address/resolver type as 'ipv4' or 'ipv6'")
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
		ch.Timeout = pingCmdConfig.timeout * 1000

		switch strings.ToLower(pingCmdConfig.ipType) {
		case "ipv4":
			ch.TargetResolver = protocheck.ResolverIPV4
		case "ipv6":
			ch.TargetResolver = protocheck.ResolverIPV6
		}

		log.WithFields(log.Fields{
			"checkId": checkId,
			"host":    host,
		}).Info("Running")
		crs, err := ch.Run()
		if err != nil {
			log.WithFields(log.Fields{
				"checkId": checkId,
				"err":     err,
			}).Warn("Ping check failed")
			return
		}

		log.WithFields(log.Fields{
			"checkId":   checkId,
			"resultSet": crs,
		}).Info("Ping check completed")
	}

}
