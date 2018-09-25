package port

import (
	"sort"

	"github.com/yittg/ving/addons"
	"github.com/yittg/ving/config"
)

func init() {
	predefinedPorts = append(wellKnownPorts, config.GetConfig().AddOns.Ports.Extra...)
	sort.Sort(predefinedPorts)

	addons.Register(NewPortAddOn())
}
