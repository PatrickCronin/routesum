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

func setupIOAndSummarize(inputPath, outputPath string) (err error) {
	var in *os.File
	if inputPath == "-" {
		in = os.Stdin
	} else {
		if in, err = os.Open(inputPath); err != nil { //nolint: gosec
			return
		}
		defer func() { //nolint: gosec // we return the error
			if cerr := in.Close(); err != nil {
				err = fmt.Errorf("failed to close input file: %w", cerr)
			}
		}()
	}

	var out *os.File
	if outputPath == "-" {
		out = os.Stdout
	} else {
		if out, err = os.Create(outputPath); err != nil {
			return
		}
		defer func() { //nolint: gosec // we return the error
			if cerr := out.Close(); err != nil {
				err = fmt.Errorf("failed to close output file: %w", cerr)
			}
		}()
	}

	return summarize(in, out)
}

func summarize(in io.Reader, out io.StringWriter) error {
	rs := routesum.NewRouteSum()
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		if err := rs.InsertFromString(string(line)); err != nil {
			return fmt.Errorf("add string: %w", err)
		}
	}

	for _, s := range rs.SummaryStrings() {
		if _, err := out.WriteString(s + "\n"); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}

	return nil
}
