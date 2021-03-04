package routesum

import (
	"fmt"
	"net"
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

// Error returns a stringified form of the error.
func (e *InvalidInputErr) Error() string {
	return fmt.Sprintf("'%s' was not understood.", e.InvalidValue)
}
