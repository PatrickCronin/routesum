# routesum

A library that summarizes a list of networks and IPs to its most succinct form.

# Synopsis

```go
strList := []string{
    "1.1.1.0",
    "1.1.1.5",
    "1.1.1.4",
    "1.1.1.2",
    "1.1.1.7",
    "207.49.18.0/24",
    "1.1.1.1",
    "1.1.1.5",
    "1.1.1.3",
    "1.1.1.6",
}

summarized := routesum.Strings(strList)

for _, s := range summarized {
    fmt.Printf("%s\n", s.String())
}
```

Will print:

```
1.1.1.0/29
207.49.18.0/24
```

# Description

This library provides two methods for summarizing a list of routes. For example,
this list of two IP addresses

```
192.0.2.0
192.0.2.1
```

can be summarized into a single network

```
192.0.2.0/31
```

The utility of this library is any place where fewer things are better, such as
when creating network firewall rules. For each packet, the fewer comparisons
required, the faster it can be discarded or routed.

The `Strings` method accepts and returns a `[]string`, while the
`NetworksAndIPs` accepts and returns both a `[]net.IPNet` network slice and a
`[]net.IP` IPs slice.

# Caveats (Maybe ToDos?)

* **IPv4-embedded IPv6 addresses**: `routesum` make heavy use of Golang's
  `net` package. The package prefers not to differentiate between an IPv4
  address (e.g. 192.0.2.1) and its IPv4-embedded IPv6 counterpart
  (i.e. ::ffff:192.0.2.1). In fact, if either were passed as input to
  `net.ParseIP`, it would be impossible to know which version had been passed by
  only by inspecting the returned `net.IP` data.

  Despite the `net` package's preference, `routesum` is able to differentiate
  between IPv4 addresses and their IPv4-embedded IPv6 counterparts. Users of
  `routesum.Strings` will not have to think about this at all, as the method
  has access to the user-supplied IP and network strings. However, users of
  `routesum.NetworksAndIPs` who care about differentiation will have to be
  careful. Under the hood, `net.IP` data is a byte slice. Slices of length 4 can
  store IPv4 addresses, and slices of length 16 can store either IPv4 or IPv6
  addresses. (Byte slices of other lengths are not valid.) In order to ensure
  that `routesum` can differentiate between IPv4 addresses and their IPv4-
  embedded counterparts, users of `NetworksAndIPs` will need to ensure that IPv4
  IPs provided to the method use the 4-byte representation. This can be
  accomplished by replacing any IPv4-representing `net.IP` variable with the
  result of calling `.To4()` on it.

* **Zero-Host Networks**: To simplify its implementation, `routesum` internally
  converts IP addresses to 0-host networks (e.g. 192.0.2.1 => 192.0.2.1/32, and
  2600:: => 2600::/128) for processing, and then converts all 0-host networks
  back to IPs when it returns its results.

* **Performance**: `routesum`'s implementation has plenty of room for
  performance improvements, and this will be clear for large inputs.

* **Sorting**: `routesum`'s output is not currently guaranteed to be sorted in
  any particular order.