package main

import (
	"os"
	"sync"
	"time"

	"github.com/yittg/ving/net"
	"github.com/yittg/ving/ui"
)

func oneLoop(ping *net.Ping, targets []string, panicErr bool) []time.Duration {
	result := make([]time.Duration, len(targets))
	wg := sync.WaitGroup{}
	for i, t := range targets {
		wg.Add(1)
		go func(idx int, addr string) {
			defer wg.Done()
			duration, e := ping.PingOnce(addr)
			if e != nil && panicErr {
				panic(e)
			}
			result[idx] = duration
		}(i, t)
	}
	wg.Wait()
	return result
}

func main() {
	ping := net.NewPing()
	ping.Start()

	targets := os.Args[1:]

	ui.Run(
		targets,
		func() ([]ui.SpItem) {
			pongs := oneLoop(ping, targets, false)
			durations := make([]ui.SpItem, len(targets))
			for idx, duration := range pongs {
				durations[idx] = ui.SpItem{Value: int(duration), Display: duration}
			}
			return durations
		},
		func() {
			ping.Stop()
		},
	)
}
