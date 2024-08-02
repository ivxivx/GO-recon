package reader

import (
	"fmt"
)

type BadFormatError struct {
	ResourceID string
	Err        error
}

func (e *BadFormatError) Error() string {
	return fmt.Sprintf("resource %s has bad format: %v", e.ResourceID, e.Err)
}

func (e *BadFormatError) Unwrap() error {
	return e.Err
}

func unwrapRootCause(err error) error {
	rootCause := err

	for {
		switch x := rootCause.(type) {
		case interface{ Unwrap() error }:
			err = x.Unwrap()
			if err == nil {
				return rootCause
			}

			rootCause = err
		default:
			return rootCause
		}
	}
}
