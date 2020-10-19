# routesum

This is a library that summarizes a list of hosts and networks into its most
consolidated form. One possible use of this is for creating router firewall
address and network groups, where the most succinct representation of such
addresses and networks is desired.

For example:

A list of IPs such as:
1.1.1.0
1.1.1.1
1.1.1.2
1.1.1.3
1.1.1.4
1.1.1.5
1.1.1.6
1.1.1.7

Could be simplified to just: 1.1.1.0/29

```go
var hosts []net.IP
var networks []net.IPNet

hosts, networks := readInput()
sumHosts, sumNets := sumroutes(hosts, networks)
```

