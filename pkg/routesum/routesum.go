// Package routesum summarizes a list of IPs and networks to its shortest form.
package routesum

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/PatrickCronin/routesum/pkg/routesum/bitvector"
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
	var bv bitvector.BitVector
	var err error

	if strings.Contains(s, "/") {
		prefix, err := netip.ParsePrefix(s)
		if err != nil {
			return fmt.Errorf("parse network: %w", err)
		}
		if !prefix.IsValid() {
			return errors.Errorf("%s is not valid CIDR", s)
		}

		ip = prefix.Addr()
		bv, err = bitvector.NewFromPrefix(prefix)
		if err != nil {
			return fmt.Errorf("convert to bits: %w", err)
		}
	} else {
		ip, err = netip.ParseAddr(s)
		if err != nil {
			return fmt.Errorf("parse IP: %w", err)
		}
		if !ip.IsValid() {
			return errors.Errorf("%s is not a valid IP", s)
		}

		bv, err = bitvector.NewFromIP(ip)
		if err != nil {
			return fmt.Errorf("convert to bits: %w", err)
		}
	}

	if ip.Is4() {
		rs.ipv4.InsertRoute(bv)
	} else {
		rs.ipv6.InsertRoute(bv)
	}

	return nil
}

// SummaryStrings returns a summary of all received routes as a string slice.
func (rs *RouteSum) SummaryStrings() ([]string, error) {
	strs := []string{}

	ipv4BitVectors := rs.ipv4.Contents()
	for _, bv := range ipv4BitVectors {
		if bv.Len() == 32 {
			ip, err := bv.ToIP(4)
			if err != nil {
				return nil, fmt.Errorf("convert to IP: %w", err)
			}

			strs = append(strs, ip.String())
		} else {
			prefix, err := bv.ToPrefix(4)
			if err != nil {
				return nil, fmt.Errorf("convert to prefix: %w", err)
			}

			strs = append(strs, prefix.String())
		}
	}

	ipv6BitVectors := rs.ipv6.Contents()
	for _, bv := range ipv6BitVectors {
		if bv.Len() == 128 {
			ip, err := bv.ToIP(16)
			if err != nil {
				return nil, fmt.Errorf("convert to IP: %w", err)
			}

			strs = append(strs, ip.String())
		} else {
			prefix, err := bv.ToPrefix(16)
			if err != nil {
				return nil, fmt.Errorf("convert to prefix: %w", err)
			}

			strs = append(strs, prefix.String())
		}
	}

	return strs, nil
}
