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
	var preRunMemStats runtime.MemStats
	if memStatsOut != nil {
		runtime.ReadMemStats(&preRunMemStats)
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

	var dataStoredMemStats runtime.MemStats
	if memStatsOut != nil {
		runtime.ReadMemStats(&dataStoredMemStats)
		logMemStatsDelta(memStatsOut, "To Store Routes", preRunMemStats, dataStoredMemStats)
	}

	for _, s := range rs.SummaryStrings() {
		if _, err := out.Write([]byte(s + "\n")); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}

	if memStatsOut != nil {
		var summaryWrittenMemStats runtime.MemStats
		runtime.ReadMemStats(&summaryWrittenMemStats)
		logMemStatsDelta(memStatsOut, "To Write Summary", dataStoredMemStats, summaryWrittenMemStats)
	}

	return nil
}

func logMemStatsDelta(w io.Writer, message string, first, second runtime.MemStats) {
	fmt.Fprintf(
		w,
		`%s
  Δ total allocated bytes: %d
  Δ mallocs:               %d
  Δ frees:                 %d
  Δ live object bytes:     %d
  Δ live objects:          %d
`,
		message,
		second.TotalAlloc-first.TotalAlloc,
		second.Mallocs-first.Mallocs,
		second.Frees-first.Frees,
		int64(second.HeapAlloc)-int64(first.HeapAlloc),
		int64(second.Mallocs-first.Mallocs)-int64(second.Frees-first.Frees),
	)
}
