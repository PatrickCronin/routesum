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

func parseArgs() args {
	fs := flag.NewFlagSet("main", flag.ExitOnError)

	inputPath := fs.String("in", "-", "File to read. Use - for STDIN.")
	outputPath := fs.String("out", "-", "File to write. Use - for STDOUT.")

	// parse the args
	err := fs.Parse(os.Args[1:])
	if errors.Is(err, flag.ErrHelp) {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s\n", os.Args[0])
		fs.PrintDefaults()
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}

	// ensure no unexpected args were provided
	if fs.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "unexpected arg: %s", fs.Arg(0))
		os.Exit(1)
	}

	cleanedInputPath := filepath.Clean(*inputPath)
	if cleanedInputPath != "-" {
		assertExtantFile("in", cleanedInputPath)
	}

	cleanedOutputPath := filepath.Clean(*outputPath)

	return args{
		inputPath:  cleanedInputPath,
		outputPath: cleanedOutputPath,
	}
}

func assertExtantFile(paramName, path string) {
	if path == "" {
		fmt.Fprintf(os.Stderr, "`%s` cannot be empty", paramName)
		os.Exit(1)
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "`%s`: %s", paramName, err.Error())
		os.Exit(1)
	} else if fileInfo.IsDir() {
		fmt.Fprintf(os.Stderr, "`%s`: %s", paramName, err.Error())
		os.Exit(1)
	}
}
