// Package main defines a program that summarizes a list of IPs and networks to its shortest form.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/PatrickCronin/routesum/pkg/routesum"
	"github.com/pkg/errors"
)

func main() {
	cpuProfile := flag.String("cpuprofile", "", "write cpu profile to file")
	memProfile := flag.String("memprofile", "", "write mem profile to file")
	flag.Parse()

	var cpuProfOut *os.File
	if *cpuProfile != "" {
		var err error
		if cpuProfOut, err = os.Create(*cpuProfile); err != nil {
			fmt.Fprint(os.Stderr, errors.Wrap(err, "create cpu profile output file").Error())
			os.Exit(1)
		}
	}

	var memProfOut *os.File
	if *memProfile != "" {
		var err error
		if memProfOut, err = os.Create(*memProfile); err != nil {
			fmt.Fprint(os.Stderr, errors.Wrap(err, "create mem profile output file").Error())
		}
	}

	if err := summarize(os.Stdin, os.Stdout, cpuProfOut, memProfOut); err != nil {
		fmt.Fprintf(os.Stderr, "summarize: %s\n", err.Error())
		os.Exit(1)
	}
}

func summarize(
	in io.Reader,
	out io.Writer,
	cpuProfOut, memProfOut io.WriteCloser,
) (retErr error) {
	// Start CPU profiling
	if cpuProfOut != nil {
		if err := pprof.StartCPUProfile(cpuProfOut); err != nil {
			return errors.Wrap(err, "start cpu profiling")
		}
		defer func() {
			if err := cpuProfOut.Close(); retErr == nil && err != nil {
				retErr = errors.Wrap(err, "close cpu profile output file")
			}
		}()
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

	// Stop CPU profiling
	if cpuProfOut != nil {
		pprof.StopCPUProfile()
	}

	// Write the allocation profile
	if memProfOut != nil {
		runtime.GC()
		if err := pprof.Lookup("allocs").WriteTo(memProfOut, 0); err != nil {
			return errors.Wrap(err, "write mem profile")
		}
		if err := memProfOut.Close(); err != nil {
			return errors.Wrap(err, "close mem profile")
		}
	}

	for _, s := range rs.SummaryStrings() {
		if _, err := out.Write([]byte(s + "\n")); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}

	return nil
}
