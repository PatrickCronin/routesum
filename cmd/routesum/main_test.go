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
	defer func() {
		err := os.Remove(in.Name())
		require.NoError(t, err, "remove temp input file")
	}()
	_, err = in.WriteString("192.0.2.0\n192.0.2.1\n")
	require.NoError(t, err, "write to temp input file")
	err = in.Close()
	require.NoError(t, err, "close temp input file")

	out, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err, "create temp output file")
	defer func() {
		err := os.Remove(out.Name())
		require.NoError(t, err, "remove temp output file")
	}()

	err = summarize(in.Name(), out.Name())
	require.NoError(t, err, "summarize does not throw an error")

	stdout := make([]byte, 50)
	n, err := out.Read(stdout)
	require.NoError(t, err, "read program stdout")
	assert.Equal(t, "192.0.2.0/31\n", string(stdout[:n]), "read expected bytes")
}
