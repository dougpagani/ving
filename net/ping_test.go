package net

import (
	"fmt"
	"sync"
	"testing"
)

func Test(t *testing.T) {
	p := NewPing()
	if err := p.Start(); err != nil {
		t.Error("start ping error", err)
	}

	targets := []string{"127.0.0.1", "example.com"}

	wg := sync.WaitGroup{}
	for _, target := range targets {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			duration, err := p.PingOnce(addr)
			if err != nil {
				fmt.Printf("ping %s error, %v\n", addr, err)
			} else {
				fmt.Printf("ping %s ttl = %v\n", addr, duration)
			}
		}(target)
	}
	wg.Wait()
}
