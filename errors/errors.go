package errors

// ErrTimeout for ping timeout error
type ErrTimeout struct {
}

func (*ErrTimeout) Error() string {
	return "timeout"
}

// ErrNoTarget represents wrong usage
type ErrNoTarget struct {

}

func (*ErrNoTarget) Error() string {
	return "no targets specified"
}
