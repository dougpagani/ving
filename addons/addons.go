package addons

// AddOn features
type AddOn interface {
	Start()

	Collect()

	Activate()

	Deactivate()

	RenderState() interface{}

	NewUI() UI
}
