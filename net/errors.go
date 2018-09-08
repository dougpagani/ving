package net

// ErrTimeout for ping timeout error
type ErrTimeout struct {
}

func (*ErrTimeout) Error() string {
	return "timeout"
}
