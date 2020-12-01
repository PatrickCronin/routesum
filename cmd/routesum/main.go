package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/PatrickCronin/routesum/pkg/routesum"
)

func main() {
	a := parseArgs()
	summarize(a.inputPath, a.outputPath)
}

func summarize(inputPath, outputPath string) {
	in, err := newInFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open input file: %s", err.Error())
		os.Exit(1)
	}
	defer in.Close()

	scanner := bufio.NewScanner(in.file)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	summarized, err := routesum.Strings(lines)
	if err != nil {
		fmt.Fprintf(os.Stderr, "summarize input: %s", err.Error())
		os.Exit(1)
	}

	out, err := newOutFile(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open output file: %s", err.Error())
		os.Exit(1)
	}
	defer out.Close()

	for _, s := range summarized {
		if _, err := out.file.WriteString(s + "\n"); err != nil {
			fmt.Fprintf(os.Stderr, "write line to output: %s", err.Error())
			os.Exit(1)
		}
	}
}
