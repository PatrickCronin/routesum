package routesum

import (
	"fmt"
	"net"
	"strings"
)

// safeRepIP is a net.IP whose representation can be counted on for certain
// properties:
// * IPv4 addresses will always be 4 bytes
// * IPv6 addresses will always be 16 bytes
// The purpose of this type is not to abstact behavior, but to be sure you can
// count on the above properties for an IP.
type safeRepIP net.IP

// newSafeRepIPFromString validates an input string as an IP and produces a
// safeRepIP.
func newSafeRepIPFromString(s string) (safeRepIP, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, fmt.Errorf("error interpreting %s as an IP", s)
	}

	if strings.Index(s, ":") != -1 { // intent is for an IPv6 IP
		return safeRepIP(ip.To16()), nil
	}

	return safeRepIP(ip.To4()), nil
}

// newSafeRepIPFromNetIP validates an input net.IP and produces a safeRepIP.
func newSafeRepIPFromNetIP(ip net.IP) (safeRepIP, error) {
	// ensure ip is valid
	if ip.To16() == nil {
		return nil, fmt.Errorf("invalid input net.IP: %#v", ip)
	}

	return safeRepIP(ip), nil
}

// safeRepIP.String() returns the stringified IP. In general, it returns
// whatever net.IP.String() would be. However it stringifies IPv4-embedded IPv6
// addresses to their IPv6 representations.
func (srIP safeRepIP) String() string {
	ip := net.IP(srIP)
	s := ip.String()

	// We trust the underlying routines for all IPv4 addresses, as well as
	// IPv6 addresses that after stringifying still look like IPv6 addresses.
	if len(srIP) == net.IPv4len || strings.Index(s, ":") != -1 {
		return s
	}

	// We stringified an IPv6 address, and now it doesn't look like an IPv6
	// address. It must be an IPv4-embedded IPv6 address. We create it's
	// IPv6-embedded IPv4 representation.
	return "::ffff:" + s
}

// safeRepNet is a net.IPNet whose representations for its IP and Mask members
// can be counted on for certain properties:
// * IPv4 networks will always have 4-byte members
// * IPv6 addresses will always have 16-byte members
// Additionally, a safeRepNet's IP will always be the IP implied by the
// network's IP and Mask combination.
// The purpose of this type is not to abstact behavior, but to be sure you can
// count on the above properties for a netowrk.
type safeRepNet net.IPNet

func newSafeRepNetFromString(s string) (*safeRepNet, error) {
	_, parsedNetwork, err := net.ParseCIDR(s)
	if err != nil {
		return nil, fmt.Errorf("error interpreting %s as a network: %w", s, err)
	}

	var network net.IPNet
	if strings.Index(s, ":") != -1 { // intent is for an IPv6 network
		network.IP = parsedNetwork.IP.To16()
		network.Mask = resizeMask(parsedNetwork.Mask, net.IPv6len*8)
	} else { // intent is for an IPv4 network
		network.IP = parsedNetwork.IP.To4()
		network.Mask = resizeMask(parsedNetwork.Mask, net.IPv4len*8)
	}

	// ensure network IP is the IP implied by the IP and Mask combination
	network.IP = network.IP.Mask(network.Mask)

	srNet := safeRepNet(network)
	return &srNet, nil
}

func newSafeRepNetFromNetIPNet(network net.IPNet) (*safeRepNet, error) {
	// ensure ip is valid
	if network.IP.To16() == nil {
		return nil, fmt.Errorf("invalid input net.IPNet.IP: %#v", network.IP)
	}

	// ensure mask is valid
	if len(network.IP) != len(network.Mask) {
		return nil, fmt.Errorf("input network IP is different length than its mask: %#v, %#v", network.IP, network.Mask)
	}

	// ensure network IP is the IP implied by the IP and Mask combination
	network.IP.Mask(network.Mask)

	srNet := safeRepNet(network)
	return &srNet, nil
}

func resizeMask(m net.IPMask, wantBits int) net.IPMask {
	onesBits, haveBits := m.Size()
	if haveBits == wantBits {
		return m
	}

	bitChange := wantBits - haveBits
	targetOnes := 0
	if onesBits+bitChange > 0 {
		targetOnes = onesBits + bitChange
	}

	return net.CIDRMask(targetOnes, wantBits)
}

// safeRepNet.String() returns the stringified network. In general this is
// whatever net.IPNet.String() would be. However, it stringifies IPv4-embedded
// IPv6 networks to their IPv6 representations.
func (srNet safeRepNet) String() string {
	network := net.IPNet(srNet)
	s := network.String()

	// We trust the underlying routines for all IPv4 addresses, as well as
	// IPv6 addresses that after stringifying still look like IPv6 addresses.
	if len(srNet.IP) == net.IPv4len || strings.Index(s, ":") != -1 {
		return s
	}

	// Otherwise, it must be an IPv4-embedded IPv6 network.
	ipStr := safeRepIP(network.IP).String()
	onesBits, _ := network.Mask.Size()
	return fmt.Sprintf("%s/%d", ipStr, onesBits)
}
