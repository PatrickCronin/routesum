package main

import (
	"fmt"
	"os"
)

type inFile struct {
	file *os.File
}

func newInFile(path string) (*inFile, error) {
	if path == "-" {
		return &inFile{file: os.Stdin}, nil
	}

	file, err := os.Open(path) // nolint: gosec
	if err != nil {
		return nil, fmt.Errorf("open file for read: %w", err)
	}

	return &inFile{file: file}, nil
}

func (in *inFile) Close() {
	if in == nil || in.file == os.Stdin {
		return
	}

	err := in.file.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to close input file")
	}
}
