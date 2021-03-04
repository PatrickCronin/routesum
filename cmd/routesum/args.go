package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type args struct {
	inputPath  string
	outputPath string
}

var errFilePathIsDir = errors.New("is a directory, not a file")

func parseArgs() (*args, error) {
	return _parseArgs(os.Args[1:])
}

func _parseArgs(theseArgs []string) (*args, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	inputPath := fs.String("in", "-", "File to read. Use - for STDIN.")
	outputPath := fs.String("out", "-", "File to write. Use - for STDOUT.")

	// parse the args
	err := fs.Parse(theseArgs)
	if errors.Is(err, flag.ErrHelp) {
		return nil, fmt.Errorf("parse: %w", err)
	}

	cleanedInputPath := filepath.Clean(*inputPath)
	if cleanedInputPath != "-" {
		if err := checkExtantFile("in", cleanedInputPath); err != nil {
			return nil, fmt.Errorf("check input file: %w", err)
		}
	}

	cleanedOutputPath := filepath.Clean(*outputPath)

	return &args{
		inputPath:  cleanedInputPath,
		outputPath: cleanedOutputPath,
	}, nil
}

func checkExtantFile(paramName, path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	} else if fileInfo.IsDir() {
		return fmt.Errorf("`%s` (`%s`) %w", paramName, path, errFilePathIsDir)
	}

	return nil
}
