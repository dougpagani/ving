package addons

// AddOn extend this utility with some useful features
// all add-ons should implements this interface
type AddOn interface {
	Desc() string

	Init(*Envoy)

	Start()

	Stop()

	Schedule()

	State() interface{}

	GetUI() UI
}
