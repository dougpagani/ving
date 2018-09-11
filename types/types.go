package types

import (
	"time"

	"github.com/yittg/ving/net/protocol"
)

// RecordHeader describes meta info of a record
type RecordHeader struct {
	ID     int
	Target *protocol.NetworkTarget
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
