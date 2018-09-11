package main

import (
	"fmt"
	"os"
	"time"

	flag "github.com/spf13/pflag"
	"github.com/yittg/ving/utils/slices"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: %s [options] target [target...]
for example: %s 127.0.0.1 192.168.0.1
             %s -i 100ms 192.168.0.1
`, slices.Repeat(os.Args[0], 3)...)
	flag.PrintDefaults()
}

type option struct {
	interval time.Duration
	timeout  time.Duration

	gateway bool

	showVersion bool
}

func (o *option) isValid() bool {
	return o.interval >= 10*time.Millisecond &&
		o.timeout >= 10*time.Millisecond
}

func parseCommandLine(opt *option) []string {
	flag.Usage = printUsage
	flag.DurationVarP(&opt.interval, "interval", "i", time.Second, "ping interval, must >=10ms")
	flag.DurationVarP(&opt.timeout, "timeout", "t", time.Second, "ping timeout, must >=10ms")
	flag.BoolVarP(&opt.gateway, "gateway", "g", false, "ping gateway");
	flag.BoolVarP(&opt.showVersion, "version", "v", false, "display the version")
	flag.Parse()

	if opt.showVersion {
		printVersion()
		os.Exit(0)
	}

	if !opt.isValid() {
		flag.Usage()
		os.Exit(1)
	}
	return flag.Args()
}
