package routesum

import (
	"errors"
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
