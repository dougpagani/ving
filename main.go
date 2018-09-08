package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/yittg/ving/net"
	"github.com/yittg/ving/types"
	"github.com/yittg/ving/ui"
)

func pingTarget(
	ping *net.Ping,
	opt *option,
	header types.RecordHeader,
	recordChan chan types.Record,
	stopChan chan bool,
) {
	if header.Target.Typ == net.Unknown {
		recordChan <- types.Record{
			RecordHeader: header,
			Successful:   false,
			ErrMsg:       header.Target.Target.(error).Error(),
			IsFatal:      true,
		}
		return
	}
	t := time.NewTicker(opt.interval)
	for {
		select {
		case <-stopChan:
			return
		case <-t.C:
			duration, e := ping.PingOnce(header.Target, opt.timeout)
			header.Rounds++
			if e != nil {
				recordChan <- types.Record{
					RecordHeader: header,
					Successful:   false,
					ErrMsg:       e.Error(),
				}
			} else {
				recordChan <- types.Record{
					RecordHeader: header,
					Successful:   true,
					Cost:         duration,
				}
			}
		}
	}
}

func main() {
	opt := parseOptions()
	targets := flag.Args()
	if len(targets) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	ping := net.NewPing()
	if err := ping.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start ping error, %v", err)
		os.Exit(2)
	}

	networkTargets := make([]*net.NetworkTarget, 0, len(targets))
	for _, t := range targets {
		networkTargets = append(networkTargets, net.ResolveTarget(t))
	}

	recordChan := make(chan types.Record, len(targets))
	stopChan := make(chan bool, 2)

	for idx, target := range networkTargets {
		header := types.RecordHeader{
			ID:     idx,
			Target: target,
		}
		go pingTarget(ping, opt, header, recordChan, stopChan)
	}

	console := ui.NewConsole(targets)
	console.Run(recordChan, func() {
		close(stopChan)
		ping.Stop()
	})
}
