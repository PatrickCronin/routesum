// Package main defines a program that summarizes a list of IPs and networks to its shortest form.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/PatrickCronin/routesum/pkg/routesum"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := summarize(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "summarize: %s\n", err.Error())
		os.Exit(1)
	}
}

func summarize(in io.Reader, out io.Writer) error {
	input := make(chan string)
	var g errgroup.Group

	// Create a new route sum and a go-routine to fill it.
	rs := routesum.NewRouteSum()
	g.Go(func() error {
		for {
			line, ok := <-input

			if !ok {
				return nil
			}

			if err := rs.InsertFromString(line); err != nil {
				return errors.Wrap(err, "add line to input")
			}
		}
	})

	// Create a go-routeine to read the input and send it to the route sum.
	g.Go(func() error {
		scanner := bufio.NewScanner(in)
		for scanner.Scan() {
			line := bytes.TrimSpace(scanner.Bytes())
			if len(line) == 0 {
				continue
			}

			input <- string(line)
		}

		close(input)

		return errors.Wrap(scanner.Err(), "read input")
	})

	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "load tries")
	}

	for s := range rs.Each() {
		if _, err := out.Write([]byte(s + "\n")); err != nil {
			return errors.Wrap(err, "write output")
		}
	}

	return nil
}
