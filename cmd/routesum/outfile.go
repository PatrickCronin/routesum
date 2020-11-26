package main

import (
	"fmt"
	"os"
)

type outFile struct {
	file *os.File
}

func newOutFile(path string) (*outFile, error) {
	if path == "-" {
		return &outFile{file: os.Stdout}, nil
	}

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("open file for write: %w", err)
	}

	return &outFile{file: file}, nil
}

func (out *outFile) Close() {
	if out == nil || out.file == os.Stdout {
		return
	}

	err := out.file.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to close output file")
	}
}
