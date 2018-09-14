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

// ErrNoTarget represents wrong usage
type ErrNoTarget struct {
}

func (*ErrNoTarget) Error() string {
	return "no targets specified"
}
