// Package bitvector implements a type and methods for creating and working with dynamically-sized,
// bit-endian, immutable bit vectors.
package bitvector

import (
	"errors"
	"fmt"
	"math/bits"
	"net/netip"
)

// BitVector is a dynamically-sized, big-endian bit vector.
type BitVector struct {
	bits   []byte
	length int
}

// New creates a BitVector from the given byte slice and length. Mostly for testing purposes.
func New(bits []byte, length int) BitVector {
	if bits == nil {
		return BitVector{bits: nil, length: length}
	}

	newBits := make([]byte, len(bits))
	copy(newBits, bits)
	return BitVector{bits: newBits, length: length}
}

// NewFromIP creates a BitVector from a netip.Addr.
func NewFromIP(ip netip.Addr) (BitVector, error) {
	b, err := ip.MarshalBinary()
	if err != nil {
		return New(nil, 0), fmt.Errorf("marshal IP to binary: %w", err)
	}

	return New(b, ip.BitLen()), nil
}

// NewFromPrefix creates a BitVector from a netip.Prefix.
func NewFromPrefix(prefix netip.Prefix) (BitVector, error) {
	b, err := prefix.Addr().MarshalBinary()
	if err != nil {
		return New(nil, 0), fmt.Errorf("marshal prefix IP to binary: %w", err)
	}

	bv := New(b, prefix.Addr().BitLen())

	// Shift off the host bits and return.
	return bv.ShiftRight(prefix.Addr().BitLen() - prefix.Bits()), nil
}

// Len returns the number of bits in the BitVector.
func (bv BitVector) Len() int {
	return bv.length
}

// Clone makes a deep copy of a BitVector.
func (bv BitVector) Clone() BitVector {
	if bv.length == 0 {
		return New(nil, 0)
	}

	numBytes := len(bv.bits)
	newBits := make([]byte, numBytes)
	copy(newBits, bv.bits)

	return New(newBits, bv.length)
}

// ShiftRight returns a new BitVector containing a copy of the receiver's bits shifted to the right
// n times.
func (bv BitVector) ShiftRight(n int) BitVector {
	// Are we shifting more bits than we have?
	if n >= bv.length {
		return New(nil, 0)
	}

	numBits := bv.length - n
	numRequiredBytes, numWholeBytes, numRemainingBits := numBytesForBits(numBits)

	newBits := make([]byte, numRequiredBytes)
	copy(newBits, bv.bits)

	// The final internal byte has bits that need zeroing out.
	if numRemainingBits > 0 {
		masks := map[int]byte{
			1: 0b10000000,
			2: 0b11000000,
			3: 0b11100000,
			4: 0b11110000,
			5: 0b11111000,
			6: 0b11111100,
			7: 0b11111110,
		}
		newBits[numWholeBytes] &= masks[numRemainingBits]
	}

	return New(newBits, numBits)
}

// ShiftLeft returns a new BitVector containing a copy of the receiver's bits shifted to the left
// n times.
func (bv BitVector) ShiftLeft(n int) BitVector {
	numBits := bv.length + n
	numRequiredBytes, _, _ := numBytesForBits(numBits)

	newBits := make([]byte, numRequiredBytes)
	copy(newBits, bv.bits)

	return New(newBits, numBits)
}

// HasPrefix determines if the receiver starts with the bits in the given BitVector.
func (bv BitVector) HasPrefix(bv2 BitVector) bool {
	if bv2.length > bv.length {
		return false
	}

	// Compare whole bytes.
	_, numWholeBytes, numRemainingBits := numBytesForBits(bv2.length)
	for i := range numWholeBytes {
		if bv.bits[i] != bv2.bits[i] {
			return false
		}
	}

	// If there's no remaining bits to compare, we're done.
	if numRemainingBits == 0 {
		return true
	}

	// Compare remaining bits.
	mask := byteMaskOnesFromLeft(numRemainingBits)
	return bv.bits[numWholeBytes]&mask == bv2.bits[numWholeBytes]&mask
}

// Prefix returns a BitVector containing a copy of the receiver's leading n bits.
func (bv BitVector) Prefix(n int) BitVector {
	return bv.ShiftRight(bv.length - n)
}

// At returns the value (0 or 1) of the requested bit of the receiver.
func (bv BitVector) At(n int) int {
	return bitValue(bv.bits[n>>3], n&7)
}

// bitValue returns the value of the byte's nth bit (0 or 1)
func bitValue(b byte, n int) int {
	return int(((b & (1 << (7 - n))) >> (7 - n)) & 1)
}

// SliceFrom returns a BitVector containing a copy of bits from the receiver starting at the given
// bit.
func (bv BitVector) SliceFrom(n int) BitVector {
	numBits := bv.length - n
	if numBits <= 0 {
		return New(nil, 0)
	}

	numRequiredBytes, _, _ := numBytesForBits(numBits)
	newBits := make([]byte, numRequiredBytes)

	_, startingByte, bitOffset := numBytesForBits(n)
	if bitOffset == 0 {
		// If we're starting on a byte boundary, we have a simple process.
		for i := range numRequiredBytes {
			newBits[i] = bv.bits[i+startingByte]
		}
	} else {
		// Otherwise, we're merging portions of two source bytes for each target byte.
		copyBitsWithOffset(bv.bits[startingByte:], bitOffset, numBits, newBits)
	}

	return New(newBits, numBits)
}

func byteMaskOnesFromRight(n int) byte {
	masks := map[int]byte{
		0: 0b00000000,
		1: 0b00000001,
		2: 0b00000011,
		3: 0b00000111,
		4: 0b00001111,
		5: 0b00011111,
		6: 0b00111111,
		7: 0b01111111,
		8: 0b11111111,
	}

	return masks[n]
}

func byteMaskOnesFromLeft(n int) byte {
	masks := map[int]byte{
		0: 0b00000000,
		1: 0b10000000,
		2: 0b11000000,
		3: 0b11100000,
		4: 0b11110000,
		5: 0b11111000,
		6: 0b11111100,
		7: 0b11111110,
		8: 0b11111111,
	}

	return masks[n]
}

func numBytesForBits(numBits int) (int, int, int) {
	numWholeBytes := numBits >> 3
	numRemainingBits := numBits & 7

	numRequiredBytes := numWholeBytes
	if numRemainingBits > 0 {
		numRequiredBytes++
	}

	return numRequiredBytes, numWholeBytes, numRemainingBits
}

// Append returns the concatenation of bv and bv2.
func (bv BitVector) Append(bv2 BitVector) BitVector {
	if bv.length == 0 {
		return bv2.Clone()
	}

	if bv2.length == 0 {
		return bv.Clone()
	}

	numBits := bv.length + bv2.length
	numRequiredBytes, _, _ := numBytesForBits(numBits)
	newBits := make([]byte, numRequiredBytes)

	// First copy over bv.
	for i := range len(bv.bits) {
		newBits[i] = bv.bits[i]
	}

	// Then add bv2.
	bitOffset := bv.length & 7
	if bitOffset == 0 {
		// We can just copy bytes from bv2
		startingByte := len(bv.bits)
		for i := range len(bv2.bits) {
			newBits[i+startingByte] = bv2.bits[i]
		}
	} else {
		// First we need to complete the last byte of newBits.
		secondBitOffset := 8 - bitOffset
		mask := byteMaskOnesFromLeft(secondBitOffset)
		newBits[len(bv.bits)-1] |= (bv2.bits[0] & mask) >> bitOffset

		// Next copy over the remainder of bv2.
		if 8-bitOffset < bv2.length {
			copyBitsWithOffset(bv2.bits, secondBitOffset, bv2.length-(secondBitOffset), newBits[len(bv.bits):])
		}
	}

	return New(newBits, numBits)
}

// copyByteswithOffset copies numBits from source (offset by bitOffet bits) to target.
// It panics if it's asked to copy more bits than are available.
func copyBitsWithOffset(source []byte, bitOffset, numBits int, target []byte) {
	// Create masks for the source lead and tail bytes
	leadMask := byteMaskOnesFromRight(8 - bitOffset)
	tailMask := byteMaskOnesFromLeft(bitOffset)
	leadLShift := bitOffset
	tailRShift := 8 - bitOffset

	// Write bytes into target by merging source lead and tail bytes.
	numBitsAvailable := len(source)*8 - bitOffset
	numWholeBytes := numBitsAvailable >> 3
	for i := range numWholeBytes {
		target[i] = source[i]&leadMask<<leadLShift + source[i+1]&tailMask>>tailRShift
	}

	// There may be some final lead bits to copy over.
	if numWholeBytes*8 < numBits {
		target[numWholeBytes] = source[numWholeBytes] & leadMask << leadLShift
	}
}

// CommonPrefixLen calculates the number of leading bits bv and bv2 have in common.
func (bv BitVector) CommonPrefixLen(bv2 BitVector) int {
	numComparableBits := min(bv.length, bv2.length)

	// First compare whole bytes.
	numComparableWholeBytes := numComparableBits >> 3
	for i := range numComparableWholeBytes {
		if x := bv.bits[i] ^ bv2.bits[i]; x != 0 {
			return (i * 8) + bits.LeadingZeros8(x)
		}
	}

	// Then compare the remaining bits.
	numRemainingBits := numComparableBits & 7
	if numRemainingBits == 0 {
		return numComparableBits
	}

	mask := byteMaskOnesFromLeft(numRemainingBits)
	if x := (bv.bits[numComparableWholeBytes] & mask) ^ (bv2.bits[numComparableWholeBytes] & mask); x != 0 {
		return (numComparableWholeBytes * 8) + bits.LeadingZeros8(x)
	}

	return (numComparableWholeBytes * 8) + numRemainingBits
}

// ErrNotAnIP is returned when the bits in a BitVector cannot be represented as an IP address.
var ErrNotAnIP = errors.New("unable to interpret BitVector as an IP")

// ToIP returns a netip.Addr from the BitVector using the given number of bytes. (4 for IPv4, 16 for
// IPv6).
func (bv BitVector) ToIP(numBytes int) (netip.Addr, error) {
	b := bv.ShiftLeft(numBytes*8 - bv.Len()).bits

	ip, ok := netip.AddrFromSlice(b)
	if !ok {
		return netip.Addr{}, ErrNotAnIP
	}

	return ip, nil
}

// ToPrefix returns a netip.Prefix from the BitVector using the given number of bytes. (4 for IPv4,
// 16 for IPv6).
func (bv BitVector) ToPrefix(numBytes int) (netip.Prefix, error) {
	ip, err := bv.ToIP(numBytes)
	if err != nil {
		return netip.Prefix{}, err
	}

	return netip.PrefixFrom(ip, bv.Len()), nil
}
