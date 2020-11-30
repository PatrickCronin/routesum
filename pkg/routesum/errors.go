package routesum

import (
	"fmt"
	"net"
)

// InvalidInputErr represents an error with ingesting or validating input
type InvalidInputErr struct {
	InvalidValue string
}

func newInvalidInputErrFromString(str string) *InvalidInputErr {
	return &InvalidInputErr{
		InvalidValue: str,
	}
}

func newInvalidInputErrFromNetIP(ip net.IP) *InvalidInputErr {
	return &InvalidInputErr{
		InvalidValue: fmt.Sprintf("%#v", ip),
	}
}

func newInvalidInputErrFromNetIPNet(net net.IPNet) *InvalidInputErr {
	return &InvalidInputErr{
		InvalidValue: fmt.Sprintf("%#v", net),
	}
}

func (e *InvalidInputErr) Error() string {
	return fmt.Sprintf("'%s' was not understood.", e.InvalidValue)
}

// As determines if a given error is (or is wrapped by) an InvalidInputErr.
func (e *InvalidInputErr) As(target error, saveTo **InvalidInputErr) bool {
	t, ok := target.(*InvalidInputErr) // nolint: errorlint
	if !ok {
		return false
	}

	*saveTo = t
	return true
}
