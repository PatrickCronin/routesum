// Package routetype implements route types for IPv4 and IPv6 addresses.
package routetype

import (
	"fmt"
	"net/netip"
	"strings"
	"unsafe"

	"github.com/PatrickCronin/routesum/pkg/routesum/uint128"
	"github.com/pkg/errors"
)

// V6 represents an IPv6 route.
type V6 struct {
	ip   *uint128.Uint128
	bits uint8
}

// ParseV6String creates a V6 route from an IPv6 IP address or prefix string.
func ParseV6String(s string) (*V6, error) {
	var baseAddr netip.Addr
	var bits uint8
	if strings.Contains(s, "/") {
		p, err := netip.ParsePrefix(s)
		if err != nil {
			return nil, errors.Wrapf(err, "parse `%s` as network", s)
		}

		baseAddr, bits = p.Addr(), uint8(p.Bits())
	} else {
		ip, err := netip.ParseAddr(s)
		if err != nil {
			return nil, fmt.Errorf("parse `%s` as IP: %w", s, err)
		}

		baseAddr, bits = ip, 128
	}

	if !baseAddr.Is6() {
		return nil, errors.Errorf("`%s` is neither an IPv6 nor an IPv4-embedded IPv6 address", s)
	}

	b, err := baseAddr.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "marshal IP to binary")
	}

	return &V6{
		ip: uint128.New(
			uint64(b[0])<<56+uint64(b[1])<<48+uint64(b[2])<<40+uint64(b[3])<<32+
				uint64(b[4])<<24+uint64(b[5])<<16+uint64(b[6])<<8+uint64(b[7]),
			uint64(b[8])<<56+uint64(b[9])<<48+uint64(b[10])<<40+uint64(b[11])<<32+
				uint64(b[12])<<24+uint64(b[13])<<16+uint64(b[14])<<8+uint64(b[15]),
		),
		bits: bits,
	}, nil
}

// MustParseV6String is similar to ParseV6String but it panics on error. Intended for tests only.
func MustParseV6String(s string) *V6 {
	r, err := ParseV6String(s)
	if err != nil {
		panic(err.Error())
	}

	return r
}

// String returns the string representation of the V6 route.
func (r *V6) String() (string, error) {
	// Note the peculiar final byte with which netip.Prefix marshals/unmarshals
	hi, lo := r.ip.Halves()

	b := []byte{
		byte((hi >> 56) & 255),
		byte((hi >> 48) & 255),
		byte((hi >> 40) & 255),
		byte((hi >> 32) & 255),
		byte((hi >> 24) & 255),
		byte((hi >> 16) & 255),
		byte((hi >> 8) & 255),
		byte(hi & 255),
		byte((lo >> 56) & 255),
		byte((lo >> 48) & 255),
		byte((lo >> 40) & 255),
		byte((lo >> 32) & 255),
		byte((lo >> 24) & 255),
		byte((lo >> 16) & 255),
		byte((lo >> 8) & 255),
		byte(lo & 255),
		r.bits,
	}

	var p netip.Prefix
	if err := p.UnmarshalBinary(b); err != nil {
		return "", fmt.Errorf("convert V6 to netip.Prefix: %w", err)
	}

	if p.IsSingleIP() {
		return p.Addr().String(), nil
	}

	return p.String(), nil
}

// Bits returns the number of relevant bits (starting from the most significant bit) in the route's
// IP.
func (r *V6) Bits() uint8 {
	return r.bits
}

// Contains determine if the route contains the given route.
func (r *V6) Contains(r2 *V6) bool {
	if r.bits > r2.bits {
		return false
	}

	return r.ip.CommonPrefixLen(r2.ip) >= r.bits
}

// CommonAncestor returns a new V6 route being the most-specific route that contains the receiver
// and the provided route.
func (r *V6) CommonAncestor(r2 *V6) *V6 {
	commonBits := min(min(r.bits, r2.bits), r.ip.CommonPrefixLen(r2.ip))

	return &V6{
		ip:   r.ip.And(uint128.Mask6(commonBits)),
		bits: commonBits,
	}
}

// NthBit returns the nth bit. The most significant bit is bit 1.
func (r *V6) NthBit(n uint8) uint8 {
	return r.ip.NthBit(n)
}

const v6RouteSize = unsafe.Sizeof(V6{}) + unsafe.Sizeof(&uint128.Uint128{}) //nolint: exhaustruct, gosec

// Size returns the memory size of a V6 route.
func (r *V6) Size() uintptr {
	return v6RouteSize
}

func min(a, b uint8) uint8 {
	if a <= b {
		return a
	}
	return b
}
