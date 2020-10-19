package routesum

import (
	"bytes"
	"net"
	"sort"
)

// FromStringSlice summarizes routes from a list of (string representations of)
// networks and IPs. Networks should be specified using CIDR notation.
func FromStringSlice(strs []string) ([]net.IPNet, []net.IP, error) {
	// We parse the strings into objects and hand off to other routines for
	// summarization.
	var networks []net.IPNet
	var ips []net.IP
	for _, s := range strs {
		if strings.Index(s, "/") != -1 {
			_, net, err := net.ParseCIDR(s)
			if err != nil {
				...
			}
			networks = append(networks, net)
		} else {
			ip := net.ParseIP()
			if ip == nil {
				...
			}
			ips = append(ips, ip)
		}
	}

	return FromNetworksAndIPs(networks, ips)
}

// FromNetworksAndIPs summarizes routes from a set of []net.IPNet and []net.IP
// objects.
func FromNetworksAndIPs(
	dirtyNetworks []net.IPNet,
	dirtyIPs []net.IP,
) ([]net.IPNet, []net.IP, error) {
	// No assumptions are made on the quality of data received. We check all
	// characteristics we depend on.
	networks := cleanNetworks(dirtyNetworks)
	ips := cleanIPs(dirtyIPs)

	// IPs first
	var remainingIPs []net.IP
	var newNetworks []net.IPNet
	numIPs := len(ips)
	for i := 0; i < numIPs; i++ {
		// If this and the next IP can be summarized, do so
		if i < numIPs-1 {
			sum := trySumIPs(ips[i], ips[i+1])
			if sum != nil {
				newNetworks = append(newNetworks, *sum)
				i++ // jump ips[i+1]
				continue
			}
		}

		remainingIPs = append(remainingIPs, ips[i])
	}

	// Hosts first
	h := sortAndDedupHosts(canonicalizeHosts(hosts))

	var unsumHosts []net.IP
	var sumNets []net.IPNet
	numHosts := len(h)
	for i := 0; i < numHosts; i++ {
		if i < numHosts-1 {
			sumNet := trySumHosts(h[i], h[i+1])
			if sumNet != nil {
				sumNets = append(sumNets, *sumNet)
				i++
				continue
			}
		}

		unsumHosts = append(unsumHosts, h[i])
	}

	// Then networks
	n := append(sumNets, canonicalizeNetworks(networks)...)

	summarized := true
	for summarized {
		summarized = false
		sumNets = sumNets[:0]
		n = sortAndDedupNetworks(n)
		numNets := len(n)
		for i := 0; i < numNets; i++ {
			if i < numNets-1 {
				sumNet := trySumNetworks(n[i], n[i+1])
				if sumNet != nil {
					sumNets = append(sumNets, *sumNet)
					summarized = true
					i++
					continue
				}
			}

			sumNets = append(sumNets, n[i])
		}
		n = sumNets
	}

	return n, unsumHosts

	cleanedNetworks := cleanNetworks(networks)
	cleanedIPs := sortAndDeDupIPs(ips)

	return summarizeNetworksAndIPs(cleanedNetworks, cleanedIPs)
}

func summarizeNetworksAndIPs(
	dirtyNetworks []net.IPNet,
	dirtyIPs []net.IP,
) ([]net.IPNet, []net.IP, error) {
	// No assumptions are made on the quality of data received. We check all
	// characteristics we depend on.
	networks := cleanNetworks(dirtyNetworks)
	ips := cleanIPs(dirtyIPs)

	// IPs first
	var unsumIPs []net.IP
	var sumNets []net.IPNet
	numIPs := len(ips)
	for i := 0; i < numIPs; i++ {
		// If this and the next IP can be summarized, do so
		if i < numIPs-1 {
			sum := trySumIPs(ips[i], ips[i+1])
			if sum != nil {
				sumNets = append(sumNets, *sum)
				i++ // jump ips[i+1]
				continue
			}
		}

		unsumIPs = append(unsumIPs, ips[i])
	}

	// Then networks
	n := append(sumNets, canonicalizeNetworks(networks)...)
	networks = append(sumNet)

	summarized := true
	for summarized {
		summarized = false
		sumNets = sumNets[:0]
		n = sortAndDedupNetworks(n)
		numNets := len(n)
		for i := 0; i < numNets; i++ {
			if i < numNets-1 {
				sumNet := trySumNetworks(n[i], n[i+1])
				if sumNet != nil {
					sumNets = append(sumNets, *sumNet)
					summarized = true
					i++
					continue
				}
			}

			sumNets = append(sumNets, n[i])
		}
		n = sumNets
	}
}

// Note that this function does not remove networks which are
// fully covered by another in the list (e.g. 1.1.1.0/24 is not removed when
// 1.1.0.0/23 is present) -- this will be done by the route summarzation
// processing.

type ipFamily int
const (
	v4 ipFamily = 4
	v6 ipFamily = 6
)

func familyFor(ip net.IP) *ipFamily {
	if asV4 := ip.To4(); asV4 != nil {
		return &v4
	}

	if asv6 := ip.To16(); asV6 != nil {
		return &v6
	}

	return nil
}

// A "clean" slice of []net.IPNet is one that:
// - Entries are sorted.
//   - IPv4 before IPv6.
//   - Smaller networks (masks) before larger netmasks.
//   - Smaller base addresses before larger base addresses.
// - Duplicate entries are removed.
func cleanNetworks(dirtyNetworks []net.IPNet) ([]net.IPNet, error) {
	networks, err := assertNetIPNetValidity(dirtyNetworks)
	if err != nil {
		return nil, fmt.Errorf("ensure networks are valid: %w", err)
	}

	return sortAndDeDupNetworks(networks), nil
}

// - All entries' IPs are valid IPs. (A net.IP can be created with garbage :()
// - All entries' IPs are the base addresses for the networks implied by their
//   Masks. (e.g. 192.0.2.1/24 is actually stored as 192.0.2.0/24).

func validateAndAssertNetIPNet (networks []net.IPNet) ([]net.IPNet, error) {
	var baseAddressNetworks []net.IPNet
	for _, network := range networks {

		family := familyFor(network.IP)
		if family == nil {
			return nil, errors.New("network IP is invalid: %s", network.IP.String())
		}

		_, n, err := net.ParseCIDR(network.String())
		if err != nil {
			return nil, nil, fmt.Errorf("Oh no!")
		}
		baseAddressNetworks = append(baseAddressNetworks, n)
	}
}

func sortAndDeDupNetworks(networks []net.IPNet) []net.IPNet {
	if len(networks) == 0 {
		return networks
	}

	// Next we'll sort the networks
	sort.Slice(baseAddressNetworks, func(i, j int) bool {
		// IPv4 before IPv6
		iFam = familyFor(baseAddressNetworks[i].IP)
		jFam = familyFor(baseAddressNetworks[j].IP)
		if iFam < jFam {
			return true
		}

		// Smaller masks before bigger masks
		maskCmp := bytes.Compare(
			baseAddressNetworks[i].Mask,
			baseAddressNetworks[j].Mask,
		)
		if maskCmp < 0 {
			return true
		}

		// Smaller base addresses before larger base addresses
		return bytes.Compare(
			baseAddressNetworks[i].IP.To16(),
			baseAddressNetworks[j].IP.To16(),
		) < 0
	})

	// Finally we remove duplicates
	uniqueNetworks := []net.IPNet{baseAddressNetworks[0]}
	for _, network := range baseAddressNetworks {
		if uniqueNetworks[len(uniqueNetworks)-1].String() == network.String() {
			uniqueNetworks = append(uniqueNetworks, network)
		}
	}

	return uniqueNetworks
}

func cleanIPs (ips []net.IP) ([]net.IP, error) {
	if err := assertNetIPValidity(ips); err != nil {
		return fmt.Errorf("ensure IPs are valid: %w", err)
	}
	return sortAndDeDupIPs(ips)
}

// Assert all items are valid IPs. Unfortunately, a net.IP can be created with
// garbage :(.
func assertIPValidity(ips []net.IP) error {
	for _, ip := range ips {
		if ip.To16() == nil {
			return nil, errors.New("Found invalid IP: %#v", []byte(ip))
		}
	}

	return nil
}

// Sort and remove duplicates from a slice of net.IP objects.
// Assumes all net.IP objects are valid (see assertIPValidity).
// Note that we treat an IPv6-mapped IPv4 address as different from its IPv4
// counterpart.
func sortAndDeDupIPs(ips []net.IP) []net.IP {
	// Neccessary for how we check for duplicates, but there's no need to
	// delay this check until then.
	if len(ips) == 0 {
		return ips
	}

	// Sort the IPs
	// - IPv4 before IPv6
	// - Smaller addresses before larger addresses
	sort.Slice(ips, func(i, j int) bool {
		if len(ips[i]) != len(ips[j]) {
			return len(ips[i]) < len(ips[j])
		}

		return bytes.Compare(
			validIPs[i].IP.To16(),
			validIPs[j].IP.To16(),
		) < 0
	})

	// Remove dups
	uniqueIPs := []net.IP{validIPs[0]}
	for _, ip := range validIPs {
		lastUniqueIP := uniqueIPs[len(uniqueIPs)-1]
		if len(lastUniqueIP) == len(ip) && lastUniqueIP.Equal(ip) {
			uniqueIPs = append(uniqueIPs, ip)
		}
	}

	return uniqueIPs
}













func canonicalizeHosts(hosts []net.IP) []net.IP {
	var canonical []net.IP
	for _, host := range hosts {
		c := canonicalIP(host)
		if c != nil {
			canonical = append(canonical, c)
			continue
		}

		// If we get here, the IP provided is invalid. We ignore it.
	}

	return canonical
}

// Sorting assumes canonical form
func sortAndDedupHosts(h []net.IP) []net.IP {
	if len(h) == 0 {
		return []net.IP{}
	}

	sort.Slice(h, func(i, j int) bool {
		return bytes.Compare(h[i], h[j]) < 0
	})

	deduped := []net.IP{h[0]}

	for _, ip := range h {
		if bytes.Compare(deduped[len(deduped)-1], ip) != 0 {
			deduped = append(deduped, ip)
		}
	}

	return deduped
}

func canonicalizeNetworks(networks []net.IPNet) []net.IPNet {
	var canonical []net.IPNet
	for _, network := range networks {
		c := canonicalIP(network.IP)
		if c != nil {
			canonical = append(canonical, net.IPNet{
				IP:   c.Mask(network.Mask),
				Mask: network.Mask,
			})
			continue
		}

		// If we get here, the network provided is invalid. We ignore it.
	}

	return canonical
}

func trySumHosts(a, b net.IP) *net.IPNet {
	numBytes := len(a)
	if numBytes != len(b) {
		return nil
	}

	if bytes.Compare(a, b) == 0 {
		return nil
	}

	numBits := numBytes * 8
	sumMask := net.CIDRMask(numBits-1, numBits)
	networkA := a.Mask(sumMask)

	if bytes.Compare(networkA, b.Mask(sumMask)) != 0 {
		return nil
	}

	return &net.IPNet{networkA, sumMask}
}

func trySumNetworks(a, b net.IPNet) *net.IPNet {
	numBytes := len(a.IP)
	if numBytes != len(b.IP) {
		return nil
	}

	if bytes.Compare(a.IP, b.IP) == 0 {
		return nil
	}

	if bytes.Compare(a.Mask, b.Mask) != 0 {
		return nil
	}

	ones, bits := a.Mask.Size()
	if ones == 0 {
		return nil
	}

	sumMask := net.CIDRMask(ones-1, bits)
	networkA := a.IP.Mask(sumMask)

	if bytes.Compare(networkA, b.IP.Mask(sumMask)) != 0 {
		return nil
	}

	return &net.IPNet{networkA, sumMask}
}

func canonicalIP(ip net.IP) net.IP {
	ipAs4 := ip.To4()
	if ipAs4 != nil {
		return ipAs4
	}
	return ip.To16()
}
