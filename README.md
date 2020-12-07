# routesum

`routesum` - summarize a list of IPs and networks to its shortest form

![Build](https://github.com/PatrickCronin/routesum/workflows/Build/badge.svg)
![golangci-lint](https://github.com/PatrickCronin/routesum/workflows/golangci-lint/badge.svg)
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

This project has utility anywhere the shortest possible for of a list of IPs and
networks is preferrable. It was initially conceived to facilitate automatic
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

## Usage

```bash
$ routesum -h
Usage of main:
  -in string
    	File to read. Use - for STDIN. (default "-")
  -out string
    	File to write. Use - for STDOUT. (default "-")
```

## Description

`routesum` is a well-behaved CLI citizen. It can take input from either a file
or from STDIN, and it can output to either a file or STDOUT.

```bash
$ routesum -in=list.txt -out=summarized.txt
$ cat input.txt | routesum > output.txt
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
Go. This project aims to maintain support for the two most recent major versions
of the Go compiler.

With this in place, simply run:

```bash
$ go get -u github.com/PatrickCronin/routesum/cmd/routesum
```

which will install `routesum` into the directory named by the `GOBIN`
environment variable, which defaults to `$GOPATH/bin` or `$HOME/go/bin` if the
`GOPATH` environment variable is not set.

# Golang Library

The `routesum` library provides two methods for users:

* `routesum.Strings()` accepts a slice of strings containing (possibly a mixture
  of) IP addresses and CIDR-formatted networks.
* `routesum.NetworksAndIPs()` accepts a slice of `net.IPNet` networks and a
  slice of `net.IP` IP addresses. Users of this methods should be sure to see
  the below-mentioned caveat on IPv4-embedded IPv6 if differentiating between
  IPv4-embedded IPv6 addresses from their IPv4 counterparts is important.

Library documentation is viewable in the code, or at
[pkg.go.dev](https://pkg.go.dev/github.com/PatrickCronin/routesum/pkg/routesum).

## Sample Code

```go
import "github.com/PatrickCronin/routesum/pkg/routesum"

ipsAndNetworks := []string{
    "198.51.100.1",
    "198.51.100.4",
    "198.51.100.5",
    "198.51.100.2/31",
    "198.51.100.6/31",
}

summarized, err := routesum.Strings(ipsAndNetworks)
if err != nil {
    ...
}

for _, s := range summarized {
    fmt.Println(s.String())
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
  will not think of `192.0.2.0` and `::ffff:192.0.2.0` as duplicates. If you
  want the two to be treated as one and the same, you'll need to ensure the data
  you provide to `routesum` in this range is consistently in one form or the
  other:

  * The `routesum` command-line program will treat IPv4-embedded IPv6 addresses
    as distinct from their IPv4 counterparts. If you don't want this, you should
    prepare the input to have all IPv4-embedded IPv6 addresses expressed either
    as IPv4 addresses or IPv6 addresses, but not a mixture of both. The data
    returned will be in the format you provided it.

  * The `routesum.Strings()` library method will treat IPv4-embedded IPv6
    addresses as distinct from their IPv4 counterparts. If you don't want this,
    you should prepare the input to have all IPv4-embedded IPv6 addresses
    expressed either as IPv4 addresses or IPv6 addresses, but not a mixture of
    both. The data returned will be in the format you provided it.

  * The `routesum.NetworksAndIPs()` library method accepts a `[]net.IPNet` and a
    `[]net.IP` as input, and takes hints from the internal representations of
    each item in each slice to distinguish IPv4-embedded IPv6 addresses from
    their IPv4 counterparts. We now need a small digression:

    The `net` package prefers not to distinguish between these two sets of IPs,
    but is capable of being cajoled into allowing us to do so. Under the hood,
    `net.IP` and `net.Mask` are `[]byte`. The package supports byte slices of
    length 4 and 16. Data with a 4-byte representation must be in the IPv4
    family, but those of 16-byte representations could have originated from
    either the IPv4 or v6 families. For example, `net.ParseIP("192.0.2.13")`
    will (currently) create a 16-byte slice to store the result, and thus if
    we're only looking at the resulting `net.IP` object, we can't tell if it was
    an IPv4 address (192.0.2.13) or its IPv4-embedded IPv6 counterpart
    (::ffff:192.0.2.13) originally. However, the `net` package offers the
    `.To4()` and `.To16()` methods to create new data using the specified
    underlying representation. To ensure that a `net.IP` is represented with 4
    bytes, take the result of calling `.To4()` on it, and similarly, to ensure
    that an IP is represented with 16 bytes, take the result of calling
    `.To16()` on it.

    If you don't want `routesum.NetworksAndIPs` to distinguish between
    IPv4-embedded IPv6 addresses and their IPv4 counterparts, you should prepare
    the input to express all IPv4 data in 4-byte representation, and all IPv6
    data in 16-byte representation, but not a mixture of both. This includes the
    `.IP` and `.Mask` fields of each `net.IPNet`, and each `net.IP`.

* **Zero-Host Networks**: To simplify its implementation, `routesum` internally
  converts IP addresses to 0-host networks (e.g. 192.0.2.1 => 192.0.2.1/32, and
  2600:: => 2600::/128) for processing, and then converts all 0-host networks
  back to IPs when it returns its results.

* **Performance**: `routesum`'s implementation has plenty of room for
  performance improvements, and this will be clear for large inputs.

* **Sorting**: `routesum`'s output is not currently guaranteed to be sorted in
  any particular order.

# Reporting Bugs and Issues

Bugs and other issues can be reported by filing an issue on our [GitHub issue
tracker](https://github.com/PatrickCronin/routesum/issues).

# Copyright and License

This software is Copyright (c) 2020 by Patrick Cronin.

This is free software, licensed under the terms of the [MIT
License](https://github.com/PatrickCronin/routesum/LICENSE.md).
