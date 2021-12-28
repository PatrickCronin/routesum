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
	if err := summarize(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "summarize: %s\n", err.Error())
		os.Exit(1)
	}
}

func summarize(in io.Reader, out io.Writer) error {
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
		if _, err := out.Write([]byte(s + "\n")); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}

	return nil
}
