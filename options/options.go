package options

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	flag "github.com/spf13/pflag"
	"github.com/yittg/ving/errors"
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

	Gateway      bool
	Trace        bool
	Ports        bool
	MorePortsStr []string
	MorePorts    []int

	Sort bool

	ShowVersion bool
}

func allPortNumber(strs ...string) ([]int, error) {
	ns := make([]int, 0, len(strs))
	for _, s := range strs {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
		if n <= 0 || n > 65535 {
			return nil, &errors.ErrInvalidPort{}
		}
		ns = append(ns, int(n))
	}
	return ns, nil
}

func (o *Option) portsValid() bool {
	for _, p := range o.MorePortsStr {
		if seg := strings.Count(p, "-"); seg > 0 {
			if seg > 1 {
				return false
			}
			pRangeS := strings.Split(p, "-")
			pRange, err := allPortNumber(pRangeS[0], pRangeS[1])
			if err != nil {
				return false
			}
			for x := pRange[0]; x <= pRange[1]; x++ {
				o.MorePorts = append(o.MorePorts, x)
			}
		} else {
			port, err := allPortNumber(p)
			if err != nil {
				return false
			}
			o.MorePorts = append(o.MorePorts, port...)
		}
	}
	return len(o.MorePorts) <= 1024
}

func (o *Option) isValid() bool {
	return o.Interval >= 10*time.Millisecond &&
		o.Timeout >= 10*time.Millisecond &&
		o.portsValid()
}

// ParseCommandLine results options and targets
func ParseCommandLine(opt *Option) []string {
	flag.Usage = printUsage
	flag.DurationVarP(&opt.Interval, "interval", "i", time.Second, "ping interval, must >=10ms")
	flag.DurationVarP(&opt.Timeout, "timeout", "t", time.Second, "ping timeout, must >=10ms")
	flag.BoolVarP(&opt.Gateway, "gateway", "g", false, "ping gateway")
	flag.BoolVarP(&opt.Trace, "trace", "T", false, "automatically traceroute the target")
	flag.BoolVarP(&opt.Ports, "ports", "", false, "automatically probe the target ports")
	flag.StringArrayVarP(&opt.MorePortsStr, "more-ports", "P", []string{},
		"ports to probe, e.g. -P 8080 -P 8082-8092, no more than #1024 ports")
	flag.BoolVarP(&opt.Sort, "sort", "", false, "sort by statistic")
	flag.BoolVarP(&opt.ShowVersion, "version", "v", false, "display the version")
	flag.Parse()

	if !opt.isValid() {
		flag.Usage()
		os.Exit(1)
	}
	return flag.Args()
}
