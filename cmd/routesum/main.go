package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/PatrickCronin/routesum/pkg/routesum"
)

func main() {
	a := parseArgs()

	err := summarize(a.inputPath, a.outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
}

func summarize(inputPath, outputPath string) error {
	in, err := newInFile(inputPath)
	if err != nil {
		return fmt.Errorf("open input file: %w", err)
	}
	defer in.Close()

	scanner := bufio.NewScanner(in.file)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	summarized, err := routesum.Strings(lines)
	if err != nil {
		return fmt.Errorf("summarize input: %w", err)
	}

	out, err := newOutFile(outputPath)
	if err != nil {
		return fmt.Errorf("open output file: %w", err)
	}
	defer out.Close()

	for _, s := range summarized {
		if _, err := out.file.WriteString(s + "\n"); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}

	return nil
}
