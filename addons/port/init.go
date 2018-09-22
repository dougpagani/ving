package port

import "github.com/yittg/ving/addons"

func init() {
	addons.Register(NewPortAddOn())
}
