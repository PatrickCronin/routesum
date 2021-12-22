// Package routesum summarizes a list of IPs and networks to its shortest form.
package routesum

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/PatrickCronin/routesum/pkg/routesum/bitslice"
	"github.com/PatrickCronin/routesum/pkg/routesum/rstree"
	"github.com/pkg/errors"
	"inet.af/netaddr"
)

// RouteSum has methods supporting route summarization of networks and hosts
type RouteSum struct {
	ipv4, ipv6 *rstree.RSTree
}

// NewRouteSum returns an initialized RouteSum object
func NewRouteSum() *RouteSum {
	rs := new(RouteSum)
	rs.ipv4 = rstree.NewRSTree()
	rs.ipv6 = rstree.NewRSTree()

	return rs
}

// InsertFromString adds either a string-formatted network or IP to the summary
func (rs *RouteSum) InsertFromString(s string) error {
	var ip netip.Addr
	var ipBits bitslice.BitSlice
	var err error

	if strings.Contains(s, "/") {
		ipPrefix, err := netip.ParsePrefix(s)
		if err != nil {
			return fmt.Errorf("parse network: %w", err)
		}
		if !ipPrefix.IsValid() {
			return errors.Errorf("%s is not valid CIDR", s)
		}

		ip = ipPrefix.Addr()
		ipBits, err = ipBitsForIPPrefix(ipPrefix)
		if err != nil {
			return err
		}
	} else {
		ip, err = netip.ParseAddr(s)
		if err != nil {
			return fmt.Errorf("parse IP: %w", err)
		}
		if !ip.IsValid() {
			return errors.Errorf("%s is not a valid IP", s)
		}

		ipBits, err = ipBitsForIP(ip)
		if err != nil {
			return err
		}
	}

	if ip.Is4() {
		rs.ipv4.InsertRoute(ipBits)
	} else {
		rs.ipv6.InsertRoute(ipBits)
	}

	return nil
}

func ipBitsForIPPrefix(ipPrefix netip.Prefix) (bitslice.BitSlice, error) {
	ipBytes, err := ipPrefix.Addr().MarshalBinary()
	if err != nil {
		return nil, errors.Wrapf(err, "express %s as bytes", ipPrefix.Addr().String())
	}

	ipBits, err := bitslice.NewFromBytes(ipBytes)
	if err != nil {
		return nil, fmt.Errorf("express %s as bits: %w", ipPrefix.Addr().String(), err)
	}

	return ipBits[:ipPrefix.Bits()], nil
}

func ipBitsForIP(ip netip.Addr) (bitslice.BitSlice, error) {
	ipBytes, err := ip.MarshalBinary()
	if err != nil {
		return nil, errors.Wrapf(err, "express %s as bytes", ip.String())
	}

	ipBits, err := bitslice.NewFromBytes(ipBytes)
	if err != nil {
		return nil, fmt.Errorf("express %s as bits: %w", ip.String(), err)
	}

	return ipBits, nil
}

// SummaryStrings returns a summary of all received routes as a string slice.
func (rs *RouteSum) SummaryStrings() []string {
	strs := []string{}

	ipv4BitSlices := rs.ipv4.Contents()
	for _, bits := range ipv4BitSlices {
		ip := ipv4FromBits(bits)

		if len(bits) == 8*4 {
			strs = append(strs, ip.String())
		} else {
			ipPrefix := netaddr.IPPrefixFrom(ip, uint8(len(bits)))
			strs = append(strs, ipPrefix.String())
		}
	}

	ipv6BitSlices := rs.ipv6.Contents()
	for _, bits := range ipv6BitSlices {
		ip := ipv6FromBits(bits)

		if len(bits) == 8*16 {
			strs = append(strs, ip.String())
		} else {
			ipPrefix := netaddr.IPPrefixFrom(ip, uint8(len(bits)))
			strs = append(strs, ipPrefix.String())
		}
	}

	return strs
}

func ipv4FromBits(bits bitslice.BitSlice) netaddr.IP {
	bytes := bits.ToBytes(4)
	byteArray := [4]byte{}
	copy(byteArray[:], bytes[0:4])
	return netaddr.IPFrom4(byteArray)
}

func ipv6FromBits(bits bitslice.BitSlice) netaddr.IP {
	bytes := bits.ToBytes(16)
	byteArray := [16]byte{}
	copy(byteArray[:], bytes[0:16])
	return netaddr.IPv6Raw(byteArray)
}

// MemUsage provides information about memory usage.
func (rs *RouteSum) MemUsage() (uint, uint, uintptr, uintptr) {
	ipv4NumInternalNodes, ipv4NumLeafNodes, ipv4InternalNodesTotalSize, ipv4LeafNodesTotalSize := rs.ipv4.MemUsage()
	ipv6NumInternalNodes, ipv6NumLeafNodes, ipv6InternalNodesTotalSize, ipv6LeafNodesTotalSize := rs.ipv6.MemUsage()
	return ipv4NumInternalNodes + ipv6NumInternalNodes,
		ipv4NumLeafNodes + ipv6NumLeafNodes,
		ipv4InternalNodesTotalSize + ipv6InternalNodesTotalSize,
		ipv4LeafNodesTotalSize + ipv6LeafNodesTotalSize
}
