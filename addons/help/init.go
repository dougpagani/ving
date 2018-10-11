package help

import "github.com/yittg/ving/addons"

func init() {
	addons.Register(NewHelp())
}
