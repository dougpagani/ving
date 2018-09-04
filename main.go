package main

import (
	"fmt"
	"os"
	"time"

	"github.com/yittg/ving/net"
	"github.com/yittg/ving/types"
	"github.com/yittg/ving/ui"
)

func pingTarget(
	ping *net.Ping,
	header types.ItemHeader,
	resChan chan interface{},
	stopChan chan bool,
) {
	t := time.NewTicker(time.Second)
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
	fmt.Printf(`%s target [target...]
    for example: %s 127.0.0.1 192.168.0.1
`, os.Args[0], os.Args[0])
}

func main() {
	targets := os.Args[1:]
	if len(targets) == 0 {
		printUsage()
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
		go pingTarget(ping, header, resChan, stopChan)
	}

	console := ui.NewConsole(targets)
	console.Run(resChan, func() {
		close(stopChan)
		ping.Stop()
	})
}
