package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/yittg/ving/net"
	"github.com/yittg/ving/types"
	"github.com/yittg/ving/ui"
	"github.com/yittg/ving/utils/slices"
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
			header.Iter += 1
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

func printUsage() {
	fmt.Fprintf(flag.CommandLine.Output(), `Usage: %s [options] target [target...]
for example: %s 127.0.0.1 192.168.0.1
             %s -i 100ms 192.168.0.1
`, slices.Repeat(os.Args[0], 3)...)
	flag.PrintDefaults()
}

type option struct {
	interval time.Duration
}

func (o *option) isValid() bool {
	return o.interval >= 10*time.Millisecond
}

func parseOptions() *option {
	flag.Usage = printUsage
	opt := option{}
	flag.DurationVar(&opt.interval, "i", time.Second, "ping interval, must >=10ms")
	flag.Parse()
	if !opt.isValid() {
		flag.Usage()
		os.Exit(1)
	}
	return &opt
}

func main() {
	opt := parseOptions()
	targets := flag.Args()
	if len(targets) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	ping := net.NewPing()
	ping.Start()

	resChan := make(chan interface{}, len(targets))
	stopChan := make(chan bool, 2)

	for idx, target := range targets {
		header := types.ItemHeader{
			Id:     idx,
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
