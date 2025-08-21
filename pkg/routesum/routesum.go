// Package routesum summarizes a list of IPs and networks to its shortest form.
package routesum

import (
	"fmt"
	"iter"
	"net/netip"
	"slices"
	"strings"

	"github.com/PatrickCronin/routesum/pkg/routesum/bitslice"
	"github.com/PatrickCronin/routesum/pkg/routesum/rstrie"
	"github.com/pkg/errors"
)

// RouteSum has methods supporting route summarization of networks and hosts
type RouteSum struct {
	ipv4, ipv6 *rstrie.RSTrie
}

// NewRouteSum returns an initialized RouteSum object
func NewRouteSum() *RouteSum {
	rs := new(RouteSum)
	rs.ipv4 = rstrie.NewRSTrie()
	rs.ipv6 = rstrie.NewRSTrie()

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
// Deprecated: Each() is preferred.
func (rs *RouteSum) SummaryStrings() []string {
	return slices.Collect(rs.Each())
}

// Each returns an iterator that returns each IP or prefix stored.
func (rs *RouteSum) Each() iter.Seq[string] {
	return func(yield func(string) bool) {
		for bits := range rs.ipv4.Each() {
			ip := ipv4FromBits(bits)

			if len(bits) == 8*4 {
				if !yield(ip.String()) {
					return
				}
			} else {
				prefix := netip.PrefixFrom(ip, len(bits))
				if !yield(prefix.String()) {
					return
				}
			}
		}

		for bits := range rs.ipv6.Each() {
			ip := ipv6FromBits(bits)

			if len(bits) == 8*16 {
				if !yield(ip.String()) {
					return
				}
			} else {
				prefix := netip.PrefixFrom(ip, len(bits))
				if !yield(prefix.String()) {
					return
				}
			}
		}
	}
}

func ipv4FromBits(bits bitslice.BitSlice) netip.Addr {
	bytes := bits.ToBytes(4)
	byteArray := [4]byte{}
	copy(byteArray[:], bytes[0:4])
	return netip.AddrFrom4(byteArray)
}

func ipv6FromBits(bits bitslice.BitSlice) netip.Addr {
	bytes := bits.ToBytes(16)
	byteArray := [16]byte{}
	copy(byteArray[:], bytes[0:16])
	return netip.AddrFrom16(byteArray)
}
