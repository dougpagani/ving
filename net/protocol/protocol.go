package protocol

// TargetType represent the type in network
type TargetType int

// Network targets types
const (
	Unknown TargetType = iota
	IP
	TCP
)
