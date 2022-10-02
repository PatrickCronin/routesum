// Package main defines a program that summarizes a list of IPs and networks to its shortest form.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/PatrickCronin/routesum/pkg/routesum"
)

func main() {
	showMemStats := flag.Bool(
		"show-mem-stats",
		false,
		"Whether or not to write memory usage stats to STDERR. (This functionity requires the use of `unsafe`, so may not be perfect.)", //nolint: lll
	)
	flag.Parse()

	var memStatsOut io.Writer
	if *showMemStats {
		memStatsOut = os.Stderr
	}

	if err := summarize(os.Stdin, os.Stdout, memStatsOut); err != nil {
		fmt.Fprintf(os.Stderr, "summarize: %s\n", err.Error())
		os.Exit(1)
	}
}

func summarize(in io.Reader, out, memStatsOut io.Writer) error {
	rs := routesum.New()
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

	if memStatsOut != nil {
		numInternalNodes, numLeafNodes, internalNodesTotalSize, leafNodesTotalSize := rs.MemUsage()
		fmt.Fprintf(memStatsOut,
			`Num internal nodes:           %d
Num leaf nodes:               %d
Size of all internal nodes:   %d
Size of all leaf nodes:       %d
Total size of data structure: %d
`,
			numInternalNodes,
			numLeafNodes,
			internalNodesTotalSize,
			leafNodesTotalSize,
			internalNodesTotalSize+leafNodesTotalSize,
		)
	}

	strs, err := rs.SummaryStrings()
	if err != nil {
		return err //nolint: wrapcheck
	}

	for _, s := range strs {
		if _, err := out.Write([]byte(s + "\n")); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}

	return nil
}
