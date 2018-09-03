package main

import (
	"os"
	"sync"

	"github.com/yittg/ving/net"
	"github.com/yittg/ving/types"
	"github.com/yittg/ving/ui"
)

func oneLoop(ping *net.Ping, targets []string) types.DataSet {
	spItems := make([]types.SpItem, len(targets))
	errItems := make([]types.ErrItem, 0, len(targets))
	errChan := make(chan types.ErrItem, len(targets))
	wg := sync.WaitGroup{}
	for i, t := range targets {
		wg.Add(1)
		go func(idx int, addr string) {
			header := types.WithId{
				Id:    addr,
				Order: idx,
			}
			defer wg.Done()
			duration, e := ping.PingOnce(addr)
			if e != nil {
				errChan <- types.ErrItem{
					WithId: header,
					Err:    e.Error(),
				}
			}
			item := types.SpItem{
				WithId:  header,
				Value:   int(duration),
				Display: duration,
			}
			if item.Value == 0 {
				item.Display = "E"
			}
			spItems[idx] = item
		}(i, t)
	}
	wg.Wait()
	for {
		select {
		case err := <-errChan:
			errItems = append(errItems, err)
		default:
			return types.DataSet{
				SpItems:  spItems,
				ErrItems: errItems,
			}
		}
	}
}

func main() {
	ping := net.NewPing()
	ping.Start()

	targets := os.Args[1:]

	ui.Run(
		targets,
		func() types.DataSet {
			return oneLoop(ping, targets)
		},
		func() {
			ping.Stop()
		},
	)
}
