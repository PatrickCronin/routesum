package routesum

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidInputErrFromString(t *testing.T) {
	e := newInvalidInputErrFromString("xyz")
	assert.Equal(t, "'xyz' was not understood.", e.Error(), "constructed error stringifies as expected")

	var iIErr *InvalidInputErr
	assert.True(t, errors.As(e, &iIErr), "unwrapped error identifies as an InvalidInputErr")
	assert.Equal(t, "'xyz' was not understood.", iIErr.Error(), "unwrapped error stringifies as expected")
}

func TestInvalidInputErrFromNetIP(t *testing.T) {
	e := newInvalidInputErrFromNetIP(net.IP([]byte{1, 2, 3}))
	assert.Equal(t, "'net.IP{0x1, 0x2, 0x3}' was not understood.", e.Error(), "constructed error stringifies as expected")

	var iIErr *InvalidInputErr
	assert.True(t, errors.As(e, &iIErr), "unwrapped error identifies as an InvalidInputErr")
	assert.Equal(t, "'net.IP{0x1, 0x2, 0x3}' was not understood.", iIErr.Error(), "unwrapped error stringifies as expected")
}

func TestInvalidInputErrFromNetIPNet(t *testing.T) {
	e := newInvalidInputErrFromNetIPNet(
		net.IPNet{
			IP:   []byte{1, 2, 3, 4},
			Mask: []byte{5, 6, 7, 8, 9},
		},
	)
	assert.Equal(t, "'net.IPNet{IP:net.IP{0x1, 0x2, 0x3, 0x4}, Mask:net.IPMask{0x5, 0x6, 0x7, 0x8, 0x9}}' was not understood.", e.Error(), "constructed error stringifies as expected")

	var iIErr *InvalidInputErr
	assert.True(t, errors.As(e, &iIErr), "coerced error identifies as an InvalidInputErr")
	assert.Equal(t, "'net.IPNet{IP:net.IP{0x1, 0x2, 0x3, 0x4}, Mask:net.IPMask{0x5, 0x6, 0x7, 0x8, 0x9}}' was not understood.", e.Error(), "unwrapped error stringifies as expected")
}
