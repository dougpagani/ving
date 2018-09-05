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
	interval time.Duration,
	header types.ItemHeader,
	resChan chan interface{},
	stopChan chan bool,
) {
	t := time.NewTicker(interval)
	for {
		select {
		case <-stopChan:
			return
		case <-t.C:
			duration, e := ping.PingOnce(header.Target)
			header.Iter++
			if e != nil {
				resChan <- types.ErrItem{
					ItemHeader: header,
					Err:        e.Error(),
				}
			} else {
				resChan <- types.SpItem{
					ItemHeader: header,
					Value:      int(duration),
					Display:    duration,
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

	resChan := make(chan interface{}, len(targets))
	stopChan := make(chan bool, 2)

	for idx, target := range targets {
		header := types.ItemHeader{
			ID:     idx,
			Target: target,
		}
		go pingTarget(ping, opt.interval, header, resChan, stopChan)
	}

	console := ui.NewConsole(targets)
	console.Run(resChan, func() {
		close(stopChan)
		ping.Stop()
	})
}
