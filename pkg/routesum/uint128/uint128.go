// Package uint128 provides a 128-bit unsigned int type. The vast majority of this type (structure
// and logic) is taken from netip/uint128. It's not direct source code redistribution, but to avoid
// problems, we follow the terms of their license, which is to reproduce their copyright. Their
// copyright is only applicable to portions of this code that are source from them.
// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package uint128

import "math/bits"

// Uint128 represents a 128-bit unsigned integer.
type Uint128 struct {
	hi uint64
	lo uint64
}

// New creates a new Uint128 from two uint64s.
func New(hi, lo uint64) *Uint128 {
	return &Uint128{hi: hi, lo: lo}
}

// Halves returns the Uint128's underlying hi and lo uint64s.
func (u *Uint128) Halves() (uint64, uint64) {
	return u.hi, u.lo
}

// Mask6 returns a uint128 bitmask with the topmost n bits of a 128-bit number.
func Mask6(n uint8) *Uint128 {
	return &Uint128{^(^uint64(0) >> n), ^uint64(0) << (128 - n)}
}

// And returns the bitwise AND of u and m (u&m).
func (u *Uint128) And(u2 *Uint128) *Uint128 {
	return &Uint128{u.hi & u2.hi, u.lo & u2.lo}
}

func u64CommonPrefixLen(a, b uint64) uint8 {
	return uint8(bits.LeadingZeros64(a ^ b))
}

// CommonPrefixLen returns the number of bits in common between two Uint128s starting from the most
// significant bit.
func (u *Uint128) CommonPrefixLen(u2 *Uint128) (n uint8) {
	if n = u64CommonPrefixLen(u.hi, u2.hi); n == 64 {
		n += u64CommonPrefixLen(u.lo, u2.lo)
	}
	return
}

// NthBit returns the nth bit. The most significant bit is bit 1.
func (u *Uint128) NthBit(n uint8) uint8 {
	if n <= 64 {
		return uint8(u.hi>>(64-n)) & 1
	}

	return uint8(u.lo>>(128-n)) & 1
}
