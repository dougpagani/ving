package port

import (
	"github.com/yittg/ving/addons"
)

func init() {
	buildPredefinedPorts()
	addons.Register(newPortAddOn())
}
