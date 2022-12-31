// Package main defines a program that summarizes a list of IPs and networks to its shortest form.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/pprof"

	"github.com/PatrickCronin/routesum/pkg/routesum"
	"github.com/pkg/errors"
)

func main() {
	cpuProfile := flag.String("cpuprofile", "", "write cpu profile to file")
	memProfile := flag.String("memprofile", "", "write mem profile to file")
	flag.Parse()

	var cpuProfOut io.Writer
	if *cpuProfile != "" {
		var err error
		if cpuProfOut, err = os.Create(*cpuProfile); err != nil {
			fmt.Fprint(os.Stderr, errors.Wrap(err, "create cpu profile output file").Error())
			os.Exit(1)
		}
	}

	var memProfOut io.WriteCloser
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
	out, cpuProfOut io.Writer,
	memProfOut io.WriteCloser,
) error {
	if cpuProfOut != nil {
		if err := pprof.StartCPUProfile(cpuProfOut); err != nil {
			return errors.Wrap(err, "start cpu profiling")
		}
		defer pprof.StopCPUProfile()
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

	if memProfOut != nil {
		if err := pprof.WriteHeapProfile(memProfOut); err != nil {
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
