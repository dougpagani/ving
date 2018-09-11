package main

import (
	"fmt"
	"os"
	"time"

	"github.com/yittg/ving/errors"
	"github.com/yittg/ving/net"
	"github.com/yittg/ving/net/protocol"
	"github.com/yittg/ving/types"
	"github.com/yittg/ving/ui"
)

func pingTarget(
	ping *net.NPing,
	opt *option,
	header types.RecordHeader,
	recordChan chan types.Record,
	stopChan chan bool,
) {
	if header.Target.Typ == protocol.Unknown {
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
				_, isTimeout := e.(*errors.ErrTimeout)
				recordChan <- types.Record{
					RecordHeader: header,
					Successful:   false,
					ErrMsg:       e.Error(),
					IsFatal:      !isTimeout,
				}
				if !isTimeout {
					return
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
	opt := option{}
	targets := parseCommandLine(&opt)
	networkTargets := make([]*protocol.NetworkTarget, 0, len(targets))
	for _, t := range targets {
		networkTargets = append(networkTargets, protocol.ResolveTarget(t))
	}
	if opt.gateway {
		networkTargets = append(networkTargets, protocol.DiscoverGatewayTarget())
	}
	if len(networkTargets) == 0 {
		printUsage()
		os.Exit(1)
	}

	recordChan := make(chan types.Record, len(networkTargets))
	stopChan := make(chan bool, 2)

	ping := net.NewPing(stopChan)
	if err := ping.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start ping error, %v", err)
		os.Exit(2)
	}

	for idx, target := range networkTargets {
		header := types.RecordHeader{
			ID:     idx,
			Target: target,
		}
		go pingTarget(ping, &opt, header, recordChan, stopChan)
	}

	console := ui.NewConsole(networkTargets)
	console.Run(recordChan, stopChan)
}
