# routesum

`routesum` - summarize a list of IPs and networks to its shortest form

![Build](https://github.com/PatrickCronin/routesum/workflows/Build/badge.svg)
![golangci-lint](https://github.com/PatrickCronin/routesum/workflows/golangci-lint/badge.svg)
[![go report](https://goreportcard.com/badge/github.com/PatrickCronin/routesum)](https://goreportcard.com/badge/github.com/PatrickCronin/routesum)
[![Coverage
Status](https://coveralls.io/repos/github/PatrickCronin/routesum/badge.svg)](https://coveralls.io/github/PatrickCronin/routesum)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/PatrickCronin/routesum/pkg/routesum)](https://pkg.go.dev/github.com/PatrickCronin/routesum/pkg/routesum)

* [Project Description](#project-description)
* [Notice of Beta Status](#notice-of-beta-status)
* [Command-Line Program](#command-line-program)
  * [Usage](#usage)
  * [Description](#description)
  * [Installation](#installation)
* [Golang Library](#golang-library)
  * [Sample Code](#sample-code)
* [Caveats](#caveats)
* [Reporting Bugs and Issues](#reporting-bugs-and-issues)
* [Copyright and License](#copyright-and-license)

# Project Description

`routesum` is both a [command-line program](#command-line-program) and a [Golang
library](#golang-library) that performs route summarization on a list of IP
addresses and CIDR-formatted networks. Route summarization is the proess of
reducing a list of IPs and networks to its shortest possible form. For example,
given the following list of IPs and networks:

    198.51.100.1
    198.51.100.4
    198.51.100.5
    198.51.100.2/31
    198.51.100.6/31

`routesum` will output:

    198.51.100.1
    198.51.100.2/31
    198.51.100.4/30

How?

* 198.51.100.4 and 198.51.100.5 summarize to 198.51.100.4/31.
* 198.51.100.4/31 and 198.51.100.6/31 summarize to 198.51.100.4/30.
* 198.51.100.1 and 198.51.100.2/31 weren't summarized, so they are just
  returned.

This project has utility anywhere the shortest possible form of a list of IPs
and networks is preferrable. It was initially conceived to facilitate automatic
daily updates to a set of firewall rules for blocking incoming connections. In
that application, the number of items on the blocklist corresponds to the
maximum number of comparisons that the firewall needs to make against every
single arriving packet. The shorter that list can be, the less work the firewall
needs to do.

# Notice of Beta Status

This software is in beta. No guarantees are made, including relating to
interface stability. Also, take note of the [caveats](#caveats) listed below.
Comments, questions or suggestions for improvements are welcome on our
[GitHub Issue Tracker](https://github.com/PatrickCronin/routesum/issues).

# Command-Line Program

## Description

`routesum` is a well-behaved CLI citizen. It takes input from STDIN, and outputs
to STDOUT.

```bash
$ routesum < infile.txt > outfile.txt
$ cat infile.txt | routesum > outfile.txt
$ routesum
192.0.2.0
192.0.2.1
^D
192.0.2.0/31
$
```

## Installation

### Binary Releases

Precompiled releases are currently available on our [Releases
page](https://github.com/PatrickCronin/routesum/releases) for the following
platforms and architectures:

* Linux (i386 and x86_64)
* macOS (x86_64 and arm64)
* Windows (i386 and x86_64)

Look for a release that ends in .tar.gz or .zip. Download the release archive
for your platform and architecture.  Uncompress the archive and you'll see an
eponymous folder. In that folder, you'll find the `routesum` program. Copy that
program to wherever you want it to live, and start using it.

### Linux Packages

Prebuilt packages are currently available on our [Releases
page](https://github.com/PatrickCronin/routesum/releases) in the following
formats:

* .deb (Ubuntu or Debian)
* .rpm (RedHat or CentOS)

On Ubuntu or Debian, use `dpkg -i /path/to/the.deb` as root. On RedHat or
CentOS, `rpm -i /path/to/the.rpm` as root. `routesum` will be installed in to
`/usr/bin/routesum`.

### Building From Source

`routesum` is written in Golang, so you'll need a reasonably recent version of
Go (1.16+). This project aims to maintain support for the two most recent major
versions of the Go compiler.

With this in place, simply run:

```bash
$ go install github.com/PatrickCronin/routesum/cmd/routesum@latest
```

which will install `routesum` into the directory named by the `GOBIN`
environment variable, which defaults to `$GOPATH/bin` or `$HOME/go/bin` if the
`GOPATH` environment variable is not set.

# Golang Library

The `routesum` library provides a type that maintains an ongoing summary of IPs
and CIDR-formatted networks as they are added:

* `rs := routesum.NewRouteSum()`

The type offers two methods:

* `rs.InsertFromString()` adds an IP or a CIDR-formatted network to its internal
  summary.
* `rs.SummaryStrings()` returns the summarized routes as a slice of strings.

Library documentation is viewable in the code, or at
[pkg.go.dev](https://pkg.go.dev/github.com/PatrickCronin/routesum/pkg/routesum).

## Sample Code

```go
import "github.com/PatrickCronin/routesum/pkg/routesum"

...

ipsAndNetworks := []string{
    "198.51.100.1",
    "198.51.100.4",
    "198.51.100.5",
    "198.51.100.2/31",
    "198.51.100.6/31",
}

rs := routesum.NewRouteSum()
for _, s := range ipsAndNetworks {
    if err := rs.InsertFromString(s); err != nil {
        ...
    }
}

summary := rs.SummaryStrings()
for _, s := range summary {
    fmt.Println(s)
}
```

Will print:

```
198.51.100.1
198.51.100.2/31
198.51.100.4/30
```

# Caveats

* **IPv4-embedded IPv6 addresses**: `routesum` treats IPv4-embedded IPv6
  addresses as distinct from their IPv4 counterparts. As an example, `routesum`
  will not think of `192.0.2.0` and `::ffff:192.0.2.0` as duplicates.

* **Zero-Host Networks**: To simplify its implementation, `routesum` internally
  converts IP addresses to 0-host networks (e.g. 192.0.2.1 => 192.0.2.1/32, and
  2600:: => 2600::/128) for processing, and then converts all 0-host networks
  back to IPs when it returns its results.

* **Sorting**: `routesum`'s output is not currently guaranteed to be sorted in
  any particular order.

# Reporting Bugs and Issues

Bugs and other issues can be reported by filing an issue on our [GitHub issue
tracker](https://github.com/PatrickCronin/routesum/issues).

# Copyright and License

This software is Copyright (c) 2020-2021 by Patrick Cronin.

This is free software, licensed under the terms of the [MIT
License](https://github.com/PatrickCronin/routesum/LICENSE.md).
