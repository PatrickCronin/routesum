package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummarize(t *testing.T) {
	inStr := "\n192.0.2.0\n192.0.2.1\n"
	in := strings.NewReader(inStr)
	var out strings.Builder

	err := summarize(in, &out)
	require.NoError(t, err, "summarize does not throw an error")

	assert.Equal(t, "192.0.2.0/31\n", out.String(), "read expected output")
}
