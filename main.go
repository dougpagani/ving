package main

import (
	"context"
	"os"

	"github.com/spf13/pflag"
	"github.com/yittg/ving/common"
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
		pflag.Usage()
		common.ErrExit("", err, 1)
	}
	ctx := context.Background()
	engine.Run(ctx)
}
