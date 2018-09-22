package options

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

// Option represents options provided
type Option struct {
	Interval time.Duration
	Timeout  time.Duration

	Gateway bool
	Trace   bool
	Ports   bool

	Sort bool

	ShowVersion bool
}

func (o *Option) isValid() bool {
	return o.Interval >= 10*time.Millisecond &&
		o.Timeout >= 10*time.Millisecond
}

// ParseCommandLine results options and targets
func ParseCommandLine(opt *Option) []string {
	flag.Usage = printUsage
	flag.DurationVarP(&opt.Interval, "interval", "i", time.Second, "ping interval, must >=10ms")
	flag.DurationVarP(&opt.Timeout, "timeout", "t", time.Second, "ping timeout, must >=10ms")
	flag.BoolVarP(&opt.Gateway, "gateway", "g", false, "ping gateway")
	flag.BoolVarP(&opt.Trace, "trace", "", false, "traceroute the target after start")
	flag.BoolVarP(&opt.Ports, "ports", "", false, "touch the target ports after start")
	flag.BoolVarP(&opt.Sort, "sort", "", false, "sort by statistic")
	flag.BoolVarP(&opt.ShowVersion, "version", "v", false, "display the version")
	flag.Parse()

	if !opt.isValid() {
		flag.Usage()
		os.Exit(1)
	}
	return flag.Args()
}
