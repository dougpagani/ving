package version

import "fmt"

const version = "0.4.1"

// PrintVersion prints current version and state
func PrintVersion() {
	versionDesc := version
	if state != "" {
		versionDesc += "-" + state
	}
	fmt.Printf("version: %s\n", versionDesc)
}
