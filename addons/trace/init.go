package trace

import "github.com/yittg/ving/addons"

func init() {
	addons.Register(NewTrace())
}
