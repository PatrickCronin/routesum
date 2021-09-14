package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/PatrickCronin/routesum/pkg/routesum"
)

func main() {
	a, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse args: %s\n", err.Error())
		os.Exit(1)
	}

	if err := setupIOAndSummarize(a.inputPath, a.outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "set up IO and summarize: %s\n", err.Error())
		os.Exit(1)
	}
}

func setupIOAndSummarize(inputPath, outputPath string) error {
	var in *os.File
	if inputPath == "-" {
		in = os.Stdin
	} else {
		var err error
		if in, err = os.Open(inputPath); err != nil { // nolint: gosec
			return fmt.Errorf("open input file for read: %w", err)
		}
		defer func() {
			if err := in.Close(); err != nil {
				fmt.Fprintln(os.Stderr, "failed to close input file")
			}
		}()
	}

	var out *os.File
	if outputPath == "-" {
		out = os.Stdout
	} else {
		var err error
		if out, err = os.Create(outputPath); err != nil {
			return fmt.Errorf("open file for write: %w", err)
		}
		defer func() {
			if err := out.Close(); err != nil {
				fmt.Fprintln(os.Stderr, "failed to close output file")
			}
		}()
	}

	if err := summarize(in, out); err != nil {
		return fmt.Errorf("summarize: %w", err)
	}

	return nil
}

func summarize(in io.Reader, out io.StringWriter) error {
	scanner := bufio.NewScanner(in)
	var lines []string
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		lines = append(lines, string(line))
	}

	summarized, err := routesum.Strings(lines)
	if err != nil {
		return fmt.Errorf("summarize input: %w", err)
	}

	for _, s := range summarized {
		if _, err := out.WriteString(s + "\n"); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}

	return nil
}
