package errors

// ErrTimeout for ping timeout error
type ErrTimeout struct {
}

func (*ErrTimeout) Error() string {
	return "timeout"
}

// ErrTTLExceed for ttl exceed
type ErrTTLExceed struct {
}

func (*ErrTTLExceed) Error() string {
	return "ttl exceed"
}

// ErrInvalidPort for invalid port
type ErrInvalidPort struct {
}

func (*ErrInvalidPort) Error() string {
	return "invalid port"
}
