package addons

import (
	"github.com/yittg/ving/net"
	"github.com/yittg/ving/net/protocol"
	"github.com/yittg/ving/options"
)

// Envoy for the ability to communicate with engine and add-ons
type Envoy struct {
	// Targets is the main targets, namely IP targets
	Targets []*protocol.NetworkTarget

	// Opt is options set when start
	Opt *options.Option

	// Ping provide major ping ability
	Ping *net.NPing
}
