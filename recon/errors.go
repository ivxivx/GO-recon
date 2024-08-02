package recon

import "fmt"

type UnexpectedTypeError struct {
	FromType any
	ToType   any
}

func (e *UnexpectedTypeError) Error() string {
	return fmt.Sprintf("cannot cast type from %T to %T", e.FromType, e.ToType)
}
