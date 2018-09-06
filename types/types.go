package types

import "time"

// RecordHeader describes meta info of a record
type RecordHeader struct {
	ID     int
	Target string
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
