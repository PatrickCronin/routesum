package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummarize(t *testing.T) {
	// create a temporary directory
	tempdir, err := ioutil.TempDir("", "routesum-test-")
	require.NoError(t, err, "create temp directory")
	defer func() {
		err := os.RemoveAll(tempdir)
		require.NoError(t, err, "remove temp directory")
	}()

	// create a test input file in the tempdir
	inPath := filepath.Clean(filepath.Join(tempdir, "in.txt"))
	err = ioutil.WriteFile(inPath, []byte("192.0.2.0\n192.0.2.1\n"), 0644)
	require.NoError(t, err, "write to temp input file")

	// run the program
	outPath := filepath.Clean(filepath.Join(tempdir, "out.txt"))
	err = setupIOAndSummarize(inPath, outPath)
	require.NoError(t, err, "summarize does not throw an error")

	// read the output
	written, err := ioutil.ReadFile(outPath)
	require.NoError(t, err, "read program output")

	assert.Equal(t, "192.0.2.0/31\n", string(written), "read expected bytes")
}
