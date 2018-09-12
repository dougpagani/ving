package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"github.com/yittg/ving/core"
	"github.com/yittg/ving/options"
)

func main() {
	opt := options.Option{}
	targets := options.ParseCommandLine(&opt)
	if opt.ShowVersion {
		printVersion()
		os.Exit(0)
	}

	engine, err := core.NewEngine(&opt, targets)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		pflag.Usage()
		os.Exit(1)
	}
	engine.Run()
}
