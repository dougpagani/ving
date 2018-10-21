package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
	_ "github.com/yittg/ving/config"
	"github.com/yittg/ving/core"
	"github.com/yittg/ving/options"
	"github.com/yittg/ving/version"
)

func main() {
	opt := options.Option{}
	targets := options.ParseCommandLine(&opt)
	if opt.ShowVersion {
		version.PrintVersion()
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
