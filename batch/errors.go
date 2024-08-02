package batch

import "fmt"

type (
	ConnectionOperationType string
	IoOperationType         string
)

const (
	ConnOpen  ConnectionOperationType = "open"
	ConnClose ConnectionOperationType = "close"

	IoOpen  IoOperationType = "open"
	IoClose IoOperationType = "close"
	IoRead  IoOperationType = "read"
	IoWrite IoOperationType = "write"
)

type IllegalArgumentError struct {
	Name  string
	Value any
}

func (e *IllegalArgumentError) Error() string {
	if e.Value == nil {
		return "illegal argument " + e.Name
	}

	return fmt.Sprintf("illegal argument %s with value %v", e.Name, e.Value)
}

type ConnectionError struct {
	Operation ConnectionOperationType
	Address   string
	Err       error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("unable to %s connection to %s: %v", e.Operation, e.Address, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

type IoError struct {
	Operation IoOperationType
	Resource  string
	Err       error
}

func (e *IoError) Error() string {
	return fmt.Sprintf("unable to %s %s: %v", e.Operation, e.Resource, e.Err)
}

func (e *IoError) Unwrap() error {
	return e.Err
}

type InvalidStatusError struct {
	StatusCode int
}

func (e *InvalidStatusError) Error() string {
	return fmt.Sprintf("invalid status %d", e.StatusCode)
}
