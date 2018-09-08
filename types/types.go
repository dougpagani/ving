package types

import (
	"time"

	"github.com/yittg/ving/net"
)

// RecordHeader describes meta info of a record
type RecordHeader struct {
	ID     int
	Target *net.NetworkTarget
	Rounds int
}

// Record records a single round result
type Record struct {
	RecordHeader

	Successful bool
	Cost       time.Duration
	ErrMsg     string
	IsFatal    bool
}
