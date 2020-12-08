package main

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultAndValidArgs(t *testing.T) {
	// default values are respected
	args := _parseArgs([]string{})
	assert.Equal(t, "-", args.inputPath, "default inputPath")
	assert.Equal(t, "-", args.outputPath, "default outputPath")

	// specified values are respected
	tempdir, err := ioutil.TempDir("", "routesum-test-")
	require.NoError(t, err, "create temp directory")
	defer func() {
		err := os.RemoveAll(tempdir)
		require.NoError(t, err, "remove temp directory")
	}()
	inPath := filepath.Join(tempdir, "in.txt")
	err = ioutil.WriteFile(inPath, []byte("192.0.2.0\n192.0.2.1\n"), 0644)
	require.NoError(t, err, "write to temp input file")

	args = _parseArgs([]string{
		"-in", inPath,
		"-out", "out.txt",
	})
	assert.Equal(t, inPath, args.inputPath, "specified inputPath")
	assert.Equal(t, "out.txt", args.outputPath, "specified outputPath")
}

func TestNonExtantInputFile(t *testing.T) {
	if os.Getenv("NON_EXTANT_INPUT_FILE") == "1" {
		_parseArgs([]string{
			"-in", "./does-not-exist.txt",
		})
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestNonExtantInputFile") // nolint: gosec
	cmd.Env = append(os.Environ(), "NON_EXTANT_INPUT_FILE=1")
	err := cmd.Run()
	var e *exec.ExitError
	if assert.True(t, errors.As(err, &e)) {
		assert.False(t, e.Success())
	}
}
