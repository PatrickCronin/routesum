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

	"golang.org/x/text/language"
	"golang.org/x/text/message"
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
		logMemStats(memStatsOut, "Before work")
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
		logMemStats(memStatsOut, "After building the summary")
	}

	for _, s := range rs.SummaryStrings() {
		if _, err := out.Write([]byte(s + "\n")); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}

	if memStatsOut != nil {
		logMemStats(memStatsOut, "After writing the summary")
	}

	return nil
}

func logMemStats(w io.Writer, message string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Fprintf(
		w,
		`%s
  HeapAlloc (excludes freed mem):   %s
  TotalAlloc (includes freed mem):  %s
  Mallocs (included freed objects): %s
  Freed objects:                    %s
`,
		message,
		formatByteCount(int64(m.HeapAlloc)),
		formatByteCount(int64(m.TotalAlloc)),
		formatNumber(m.Mallocs),
		formatNumber(m.Frees),
	)
}

func formatByteCount(b int64) string {
	const unit = 1024

	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div := int64(unit)
	exp := 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func formatNumber(n interface{}) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", n)
}
