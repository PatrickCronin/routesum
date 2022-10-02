package routetype

import (
	"fmt"
	"math/bits"
	"net/netip"
	"strings"
	"unsafe"

	"github.com/pkg/errors"
)

// V4 represents an IPv4 route.
type V4 struct {
	ip   uint32
	bits uint8
}

// ParseV4String creates a V4 route from an IPv4 IP address or prefix string.
func ParseV4String(s string) (*V4, error) {
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

		baseAddr, bits = ip, 32
	}

	if !baseAddr.Is4() {
		return nil, errors.Errorf("`%s` is not an IPv4 address", s)
	}

	b, err := baseAddr.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "marshal IP to binary")
	}

	return &V4{
		ip:   uint32(b[0])<<24 + uint32(b[1])<<16 + uint32(b[2])<<8 + uint32(b[3]),
		bits: bits,
	}, nil
}

// MustParseV4String is similar to ParseV4String but it panics on error. Intended for tests only.
func MustParseV4String(s string) *V4 {
	r, err := ParseV4String(s)
	if err != nil {
		panic(err.Error())
	}

	return r
}

// String returns the string representation of the V4 route.
func (r *V4) String() (string, error) {
	// Note the peculiar final byte with which netip.Prefix marshals/unmarshals
	b := []byte{
		byte((r.ip >> 24) & 255),
		byte((r.ip >> 16) & 255),
		byte((r.ip >> 8) & 255),
		byte(r.ip & 255),
		r.bits,
	}

	var p netip.Prefix
	if err := p.UnmarshalBinary(b); err != nil {
		return "", fmt.Errorf("convert V4 to netip.Prefix: %w", err)
	}

	if p.IsSingleIP() {
		return p.Addr().String(), nil
	}

	return p.String(), nil
}

// Bits returns the number of relevant bits (starting from the most significant bit) in the route's
// IP.
func (r *V4) Bits() uint8 {
	return r.bits
}

// Contains determine if the route contains the given route.
func (r *V4) Contains(r2 *V4) bool {
	if r.bits > r2.bits {
		return false
	}

	return (r.ip^r2.ip)>>(32-r.bits) == 0
}

// CommonAncestor returns a new V4 route being the most-specific route that contains the receiver
// and the provided route.
func (r *V4) CommonAncestor(r2 *V4) *V4 {
	commonBits := min(min(r.bits, r2.bits), uint8(bits.LeadingZeros32(r2.ip^r.ip)))
	return &V4{
		ip:   r.ip & ^((1 << (32 - commonBits)) - 1),
		bits: commonBits,
	}
}

// NthBit returns the nth bit. The most significant bit is bit 1.
func (r *V4) NthBit(n uint8) uint8 {
	return uint8((r.ip >> (32 - n)) & 1)
}

const v4RouteSize = unsafe.Sizeof(V4{}) //nolint: exhaustruct, gosec

// Size returns the memory size of a V4 route.
func (r *V4) Size() uintptr {
	return v4RouteSize
}
