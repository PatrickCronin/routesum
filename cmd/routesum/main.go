package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/PatrickCronin/routesum/pkg/routesum"
)

func main() {
	showMemStats := flag.Bool(
		"show-mem-stats",
		false,
		"Whether or not to write memory usage stats to STDERR.",
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
	if memStatsOut != nil {
		logMemStats(memStatsOut, "Before Summarize")
	}

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

	if memStatsOut != nil {
		logMemStats(memStatsOut, "After Summarize")
	}

	for _, s := range rs.SummaryStrings() {
		if _, err := out.Write([]byte(s + "\n")); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}

	if memStatsOut != nil {
		logMemStats(memStatsOut, "After Writing")
	}

	return nil
}

func logMemStats(w io.Writer, message string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Fprintf(
		w,
		`%s
  HeapAlloc (excludes freed mem):   %d
  TotalAlloc (includes freed mem):  %d
  Mallocs (included freed objects): %d
  Freed objects:                    %d
`,
		message,
		m.HeapAlloc,
		m.TotalAlloc,
		m.Mallocs,
		m.Frees,
	)
}
