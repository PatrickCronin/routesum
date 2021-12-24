package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArgs(t *testing.T) {
	// default values are respected
	args, err := _parseArgs([]string{})
	require.NoError(t, err, "parsing args")
	assert.Equal(t, "-", args.inputPath, "default inputPath")
	assert.Equal(t, "-", args.outputPath, "default outputPath")

	// specified values are respected
	tempdir, err := os.MkdirTemp("", "routesum-test-")
	require.NoError(t, err, "create temp directory")
	defer func() {
		err := os.RemoveAll(tempdir)
		require.NoError(t, err, "remove temp directory")
	}()
	inPath := filepath.Join(tempdir, "in.txt")
	err = os.WriteFile(inPath, []byte("192.0.2.0\n192.0.2.1\n"), 0o644)
	require.NoError(t, err, "write to temp input file")

	args, err = _parseArgs([]string{
		"-in", inPath,
		"-out", "out.txt",
	})
	require.NoError(t, err, "parsing args")
	assert.Equal(t, inPath, args.inputPath, "specified inputPath")
	assert.Equal(t, "out.txt", args.outputPath, "specified outputPath")

	// non-extant input file throws error
	_, err = _parseArgs([]string{"-in", "./does-not-exist.txt"})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "check input file", "expected error on non-extant input file")
	}
}
