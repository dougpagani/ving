package addons

import "context"

// AddOn extend this utility with some useful features
// all add-ons should implements this interface
type AddOn interface {
	Desc() string

	Init(*Envoy)

	Start(context.Context)

	Schedule()

	State() interface{}

	GetUI() UI
}
