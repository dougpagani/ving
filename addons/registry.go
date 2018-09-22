package addons

// All of add-ons
var All []AddOn

// Register add-on
func Register(addOn AddOn) {
	All = append(All, addOn)
}
