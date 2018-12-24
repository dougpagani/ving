package common

import (
	"fmt"
	"os"
)

func ErrExit(msg string, err error, exitCode int) {
	if msg == "" {
		msg = "error"
	}
	_, _ = fmt.Fprintf(os.Stderr, "%s, cause:%s", msg, err)
	os.Exit(exitCode)
}
