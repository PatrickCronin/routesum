package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleUsage(t *testing.T) {
	in, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err, "create temp input file")
	defer os.Remove(in.Name())
	in.WriteString("1.1.1.0\n1.1.1.1\n")
	err = in.Close()
	require.NoError(t, err, "close temp input file")

	out, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err, "create temp output file")
	defer os.Remove(out.Name())

	summarize(in.Name(), out.Name())

	stdout := make([]byte, 50)
	n, err := out.Read(stdout)
	require.NoError(t, err, "read program stdout")
	assert.Equal(t, "1.1.1.0/31\n", string(stdout[:n]), "read expected bytes")
}
