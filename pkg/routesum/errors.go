package routesum

import (
	"fmt"
)

// InvalidInputErr represents an error ingesting or validating input
type InvalidInputErr struct {
	InvalidValue string
}

func newInvalidInputErrFromString(str string) *InvalidInputErr {
	return &InvalidInputErr{
		InvalidValue: str,
	}
}

// Error returns a stringified form of the error.
func (e *InvalidInputErr) Error() string {
	return fmt.Sprintf("'%s' was not understood.", e.InvalidValue)
}
