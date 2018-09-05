package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/yittg/ving/utils/slices"
)

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
