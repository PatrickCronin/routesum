// Package bitslice provides a slice of bits
package bitslice

// BitSlice is a slice of zeros and ones
type BitSlice []byte

// NewFromBytes creates a BitSlice from a byte slice
func NewFromBytes(bytes []byte) (BitSlice, error) {
	bits := make(BitSlice, 0, len(bytes)*8)

	// Convert each byte to its binary representation as bits
	for _, b := range bytes {
		bits = append(
			bits,
			b&0b10000000>>7,
			b&0b01000000>>6,
			b&0b00100000>>5,
			b&0b00010000>>4,
			b&0b00001000>>3,
			b&0b00000100>>2,
			b&0b00000010>>1,
			b&0b00000001,
		)
	}

	return bits, nil
}

// ToBytes packages a BitSlice as a byte slice of length numBytes
func (b BitSlice) ToBytes(numBytes int) []byte {
	// Ensure we're working with enough bits for the requested bytes
	completeBits := make([]byte, numBytes*8)
	copy(completeBits, b)

	bytes := make([]byte, 0, numBytes)
	for i := 0; i < numBytes; i++ {
		bytes = append(
			bytes,
			completeBits[i*8]<<7+
				completeBits[i*8+1]<<6+
				completeBits[i*8+2]<<5+
				completeBits[i*8+3]<<4+
				completeBits[i*8+4]<<3+
				completeBits[i*8+5]<<2+
				completeBits[i*8+6]<<1+
				completeBits[i*8+7],
		)
	}

	return bytes
}
