package bitvector

import (
	"maps"
	"net/netip"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInputOutputForIP(t *testing.T) {
	tests := []struct {
		name     string
		val      netip.Addr
		expected BitVector
	}{
		{
			name:     "IPv4 addr",
			val:      netip.MustParseAddr("192.0.2.37"),
			expected: New([]byte{192, 0, 2, 37}, 32),
		},
		{
			name: "IPv6-embedded IPv4 addr",
			val:  netip.MustParseAddr("::ffff:192.0.2.37"),
			expected: New(
				[]byte{
					0, 0, 0, 0,
					0, 0, 0, 0,
					0, 0, 0xff, 0xff,
					192, 0, 2, 37,
				},
				128,
			),
		},
		{
			name: "IPv6 addr",
			val:  netip.MustParseAddr("2001:db8::1"),
			expected: New(
				[]byte{
					0x20, 0x1, 0xd, 0xb8,
					0, 0, 0, 0,
					0, 0, 0, 0,
					0, 0, 0, 1,
				},
				128,
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bv, err := NewFromIP(test.val)
			require.NoError(t, err)

			assert.Equal(t, test.expected, bv)

			gotIP, err := bv.ToIP(test.val.BitLen() / 8)
			require.NoError(t, err)

			assert.Equal(t, test.val, gotIP)
		})
	}
}

func TestInputOutputForPrefix(t *testing.T) {
	tests := []struct {
		name     string
		val      netip.Prefix
		expected BitVector
	}{
		{
			name:     "IPv4 prefix",
			val:      netip.MustParsePrefix("192.0.2.0/28"),
			expected: New([]byte{192, 0, 2, 0}, 28),
		},
		{
			name: "IPv6-embedded IPv4 prefix",
			val:  netip.MustParsePrefix("::ffff:192.0.2.0/127"),
			expected: New(
				[]byte{
					0, 0, 0, 0,
					0, 0, 0, 0,
					0, 0, 0xff, 0xff,
					192, 0, 2, 0,
				},
				127,
			),
		},
		{
			name:     "IPv6 prefix",
			val:      netip.MustParsePrefix("2001:db8::/29"),
			expected: New([]byte{0x20, 0x1, 0xd, 0xb8}, 29),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bv, err := NewFromPrefix(test.val)
			require.NoError(t, err)

			assert.Equal(t, test.expected, bv)

			gotPrefix, err := bv.ToPrefix(test.val.Addr().BitLen() / 8)
			require.NoError(t, err)

			assert.Equal(t, test.val, gotPrefix)
		})
	}
}

func TestShiftRight(t *testing.T) { //nolint: funlen // we're testing a lot
	tests := []struct {
		name     string
		bv       BitVector
		numBits  int
		expected BitVector
	}{
		{
			name:     "identity",
			bv:       New([]byte{0b11111111, 0b11110000}, 12),
			numBits:  0,
			expected: New([]byte{0b11111111, 0b11110000}, 12),
		},
		{
			name:     "shift more than we have",
			bv:       New([]byte{0b11111111, 0b11110000}, 12),
			numBits:  12,
			expected: New(nil, 0),
		},
		{
			name:     "shift one bit",
			bv:       New([]byte{0b10101010, 0b10100000}, 12),
			numBits:  1,
			expected: New([]byte{0b10101010, 0b10100000}, 11),
		},
		{
			name:     "shift two bits",
			bv:       New([]byte{0b10101010, 0b10100000}, 12),
			numBits:  2,
			expected: New([]byte{0b10101010, 0b10000000}, 10),
		},
		{
			name:     "shift four bits, disappearing one byte",
			bv:       New([]byte{0b10101010, 0b10100000}, 12),
			numBits:  4,
			expected: New([]byte{0b10101010}, 8),
		},
		{
			name:     "shift a byte, no bits",
			bv:       New([]byte{0b10101010, 0b10100000}, 12),
			numBits:  8,
			expected: New([]byte{0b10100000}, 4),
		},
		{
			name:     "shift a byte and a bit",
			bv:       New([]byte{0b10101010, 0b10100000}, 12),
			numBits:  9,
			expected: New([]byte{0b10100000}, 3),
		},
		{
			name:     "mask off one bit",
			bv:       New([]byte{0b11111111}, 8),
			numBits:  1,
			expected: New([]byte{0b11111110}, 7),
		},
		{
			name:     "mask off two bits",
			bv:       New([]byte{0b11111111}, 8),
			numBits:  2,
			expected: New([]byte{0b11111100}, 6),
		},
		{
			name:     "mask off three bits",
			bv:       New([]byte{0b11111111}, 8),
			numBits:  3,
			expected: New([]byte{0b11111000}, 5),
		},
		{
			name:     "mask off four bits",
			bv:       New([]byte{0b11111111}, 8),
			numBits:  4,
			expected: New([]byte{0b11110000}, 4),
		},
		{
			name:     "mask off five bits",
			bv:       New([]byte{0b11111111}, 8),
			numBits:  5,
			expected: New([]byte{0b11100000}, 3),
		},
		{
			name:     "mask off six bits",
			bv:       New([]byte{0b11111111}, 8),
			numBits:  6,
			expected: New([]byte{0b11000000}, 2),
		},
		{
			name:     "mask off seven bits",
			bv:       New([]byte{0b11111111}, 8),
			numBits:  7,
			expected: New([]byte{0b10000000}, 1),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.bv.ShiftRight(test.numBits)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestShiftLeft(t *testing.T) {
	tests := []struct {
		name     string
		bv       BitVector
		numBits  int
		expected BitVector
	}{
		{
			name:     "identity",
			bv:       New([]byte{0b11111111, 0b11110000}, 12),
			numBits:  0,
			expected: New([]byte{0b11111111, 0b11110000}, 12),
		},
		{
			name:     "shift one bit",
			bv:       New([]byte{0b10101010, 0b10100000}, 12),
			numBits:  1,
			expected: New([]byte{0b10101010, 0b10100000}, 13),
		},
		{
			name:     "shift two bits",
			bv:       New([]byte{0b10101010, 0b10100000}, 12),
			numBits:  2,
			expected: New([]byte{0b10101010, 0b10100000}, 14),
		},
		{
			name:     "shift four bits, filling the last byte",
			bv:       New([]byte{0b10101010, 0b10100000}, 12),
			numBits:  4,
			expected: New([]byte{0b10101010, 0b10100000}, 16),
		},
		{
			name:     "shift a byte, no bits",
			bv:       New([]byte{0b10101010, 0b10100000}, 12),
			numBits:  8,
			expected: New([]byte{0b10101010, 0b10100000, 0b00000000}, 20),
		},
		{
			name:     "shift a byte and a bit",
			bv:       New([]byte{0b10101010, 0b10100000}, 12),
			numBits:  9,
			expected: New([]byte{0b10101010, 0b10100000, 0b00000000}, 21),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.bv.ShiftLeft(test.numBits)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestHasPrefix(t *testing.T) {
	tests := []struct {
		name     string
		bv1      BitVector
		bv2      BitVector
		expected bool
	}{
		{
			name:     "identity",
			bv1:      New([]byte{0b11111111, 0b11110000}, 12),
			bv2:      New([]byte{0b11111111, 0b11110000}, 12),
			expected: true,
		},
		{
			name:     "shorter - same",
			bv1:      New([]byte{0b11111111, 0b11110000}, 12),
			bv2:      New([]byte{0b11111111, 0b11100000}, 11),
			expected: true,
		},
		{
			name:     "shorter - not same on bits",
			bv1:      New([]byte{0b11111111, 0b11110000}, 12),
			bv2:      New([]byte{0b11111111, 0b00000000}, 11),
			expected: false,
		},
		{
			name:     "shorter - not same on bytes",
			bv1:      New([]byte{0b11111111, 0b11110000}, 12),
			bv2:      New([]byte{0b00010010, 0b00000000}, 11),
			expected: false,
		},
		{
			name:     "shorter - on byte boundary",
			bv1:      New([]byte{0b11111111, 0b11110000}, 12),
			bv2:      New([]byte{0b11111111}, 8),
			expected: true,
		},
		{
			name:     "one byte shorter",
			bv1:      New([]byte{0b11111111, 0b11110000}, 12),
			bv2:      New([]byte{0b11111110}, 7),
			expected: true,
		},
		{
			name:     "longer",
			bv1:      New([]byte{0b11111111, 0b11110000}, 12),
			bv2:      New([]byte{0b11111111, 0b11111000}, 13),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.bv1.HasPrefix(test.bv2))
		})
	}
}

func TestAt(t *testing.T) {
	bv := New([]byte{192, 0, 2, 36}, 31)

	expectedBits := []int{
		1, 1, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 1, 0,
		0, 0, 1, 0, 0, 1, 0,
	}

	assert.Equal(t, 31, bv.length)

	for i := range 31 {
		assert.Equal(t, expectedBits[i], bv.At(i), "bit "+strconv.Itoa(i))
	}
}

func TestSliceFrom(t *testing.T) {
	bv := New([]byte{0b10101010, 0b10101010, 0b10100000}, 20)

	tests := map[int]BitVector{
		0:  bv,
		1:  New([]byte{0b01010101, 0b01010101, 0b01000000}, 19),
		2:  New([]byte{0b10101010, 0b10101010, 0b10000000}, 18),
		3:  New([]byte{0b01010101, 0b01010101, 0b00000000}, 17),
		4:  New([]byte{0b10101010, 0b10101010}, 16),
		5:  New([]byte{0b01010101, 0b01010100}, 15),
		6:  New([]byte{0b10101010, 0b10101000}, 14),
		7:  New([]byte{0b01010101, 0b01010000}, 13),
		8:  New([]byte{0b10101010, 0b10100000}, 12),
		9:  New([]byte{0b01010101, 0b01000000}, 11),
		10: New([]byte{0b10101010, 0b10000000}, 10),
		11: New([]byte{0b01010101, 0b00000000}, 9),
		12: New([]byte{0b10101010}, 8),
		13: New([]byte{0b01010100}, 7),
		14: New([]byte{0b10101000}, 6),
		15: New([]byte{0b01010000}, 5),
		16: New([]byte{0b10100000}, 4),
		17: New([]byte{0b01000000}, 3),
		18: New([]byte{0b10000000}, 2),
		19: New([]byte{0b00000000}, 1),
		20: New(nil, 0),
	}

	sortedNumBits := slices.Collect(maps.Keys(tests))
	slices.Sort(sortedNumBits)

	for _, numBits := range sortedNumBits {
		t.Run(strconv.Itoa(numBits)+" bits", func(t *testing.T) {
			assert.Equal(t, tests[numBits], bv.SliceFrom(numBits))
		})
	}
}

func TestAppend(t *testing.T) { //nolint: funlen // We're testing a lot.
	tests := []struct {
		name               string
		bv1, bv2, expected BitVector
	}{
		{
			name:     "empty appended to empty",
			bv1:      New(nil, 0),
			bv2:      New(nil, 0),
			expected: New(nil, 0),
		},
		{
			name:     "empty appended to not empty",
			bv1:      New(nil, 0),
			bv2:      BitVector{bits: []byte{0b10101010, 0b10000000}, length: 10},
			expected: BitVector{bits: []byte{0b10101010, 0b10000000}, length: 10},
		},
		{
			name:     "not empty appended to empty",
			bv1:      BitVector{bits: []byte{0b10101111}, length: 8},
			bv2:      New(nil, 0),
			expected: BitVector{bits: []byte{0b10101111}, length: 8},
		},
		{
			name:     "some bits and some bits, result less than a byte",
			bv1:      BitVector{bits: []byte{0b10100000}, length: 3},
			bv2:      BitVector{bits: []byte{0b11000000}, length: 2},
			expected: BitVector{bits: []byte{0b10111000}, length: 5},
		},
		{
			name:     "some bits and some bits, result exactly a byte",
			bv1:      BitVector{bits: []byte{0b10100000}, length: 3},
			bv2:      BitVector{bits: []byte{0b10111000}, length: 5},
			expected: BitVector{bits: []byte{0b10110111}, length: 8},
		},
		{
			name:     "some bits and some bits, result more than a byte",
			bv1:      BitVector{bits: []byte{0b10101010}, length: 7},
			bv2:      BitVector{bits: []byte{0b10111000}, length: 5},
			expected: BitVector{bits: []byte{0b10101011, 0b01110000}, length: 12},
		},
		{
			name:     "some bits and a byte",
			bv1:      BitVector{bits: []byte{0b10100000}, length: 3},
			bv2:      BitVector{bits: []byte{0b01010101}, length: 8},
			expected: BitVector{bits: []byte{0b10101010, 0b10100000}, length: 11},
		},
		{
			name:     "some bits and a byte and some bits, resulting in a byte and some bits",
			bv1:      BitVector{bits: []byte{0b10100000}, length: 3},
			bv2:      BitVector{bits: []byte{0b01010101, 0b01110000}, length: 12},
			expected: BitVector{bits: []byte{0b10101010, 0b10101110}, length: 15},
		},
		{
			name:     "some bits and a byte and some bits, resulting in a two bytes",
			bv1:      BitVector{bits: []byte{0b10100100}, length: 6},
			bv2:      BitVector{bits: []byte{0b10111000, 0b10000000}, length: 10},
			expected: BitVector{bits: []byte{0b10100110, 0b11100010}, length: 16},
		},
		{
			name:     "some bits and some bytes",
			bv1:      BitVector{bits: []byte{0b11010000}, length: 4},
			bv2:      BitVector{bits: []byte{0b00100010, 0b00001000, 0b10000000}, length: 20},
			expected: BitVector{bits: []byte{0b11010010, 0b00100000, 0b10001000}, length: 24},
		},
		{
			name:     "a byte and some bits",
			bv1:      BitVector{bits: []byte{0b01000101}, length: 8},
			bv2:      BitVector{bits: []byte{0b10000000}, length: 2},
			expected: BitVector{bits: []byte{0b01000101, 0b10000000}, length: 10},
		},
		{
			name:     "a byte and a byte",
			bv1:      BitVector{bits: []byte{0b01101100}, length: 8},
			bv2:      BitVector{bits: []byte{0b00110101}, length: 8},
			expected: BitVector{bits: []byte{0b01101100, 0b00110101}, length: 16},
		},
		{
			name:     "a byte and some bytes and bits",
			bv1:      BitVector{bits: []byte{0b10010001}, length: 8},
			bv2:      BitVector{bits: []byte{0b01000111, 0b00110100, 0b11011110}, length: 23},
			expected: BitVector{bits: []byte{0b10010001, 0b01000111, 0b00110100, 0b11011110}, length: 31},
		},
		{
			name:     "a byte and some bits and some bits, result is less than two bytes",
			bv1:      BitVector{bits: []byte{0b01111110, 0b01000000}, length: 10},
			bv2:      BitVector{bits: []byte{0b00000000}, length: 3},
			expected: BitVector{bits: []byte{0b01111110, 0b01000000}, length: 13},
		},
		{
			name:     "a byte and some bits and some bits, result is two bytes",
			bv1:      BitVector{bits: []byte{0b01111110, 0b01000000}, length: 10},
			bv2:      BitVector{bits: []byte{0b01010100}, length: 6},
			expected: BitVector{bits: []byte{0b01111110, 0b01010101}, length: 16},
		},
		{
			name:     "a byte and some bits and some bits, result is more than two bytes",
			bv1:      BitVector{bits: []byte{0b01111110, 0b01000000}, length: 10},
			bv2:      BitVector{bits: []byte{0b00001110}, length: 7},
			expected: BitVector{bits: []byte{0b01111110, 0b01000011, 0b10000000}, length: 17},
		},
		{
			name:     "a byte and some bits and a byte",
			bv1:      BitVector{bits: []byte{0b11001010, 0b10100000}, length: 11},
			bv2:      BitVector{bits: []byte{0b00000110}, length: 8},
			expected: BitVector{bits: []byte{0b11001010, 0b10100000, 0b11000000}, length: 19},
		},
		{
			name:     "a byte and some bits and a byte and some bits, result is less than three bytes",
			bv1:      BitVector{bits: []byte{0b00010110, 0b01000000}, length: 10},
			bv2:      BitVector{bits: []byte{0b01100010, 0b11000000}, length: 11},
			expected: BitVector{bits: []byte{0b00010110, 0b01011000, 0b10110000}, length: 21},
		},
		{
			name:     "a byte and some bits and a byte and some bits, result is three bytes",
			bv1:      BitVector{bits: []byte{0b10011111, 0b10000100}, length: 15},
			bv2:      BitVector{bits: []byte{0b11010110, 0b00000000}, length: 9},
			expected: BitVector{bits: []byte{0b10011111, 0b10000101, 0b10101100}, length: 24},
		},
		{
			name:     "a byte and some bits and a byte and some bits, result is more than three bytes",
			bv1:      BitVector{bits: []byte{0b00010110, 0b00110010}, length: 15},
			bv2:      BitVector{bits: []byte{0b01100010, 0b11000000}, length: 14},
			expected: BitVector{bits: []byte{0b00010110, 0b00110010, 0b11000101, 0b10000000}, length: 29},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.bv1.Append(test.bv2)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestCommonPrefixLen(t *testing.T) {
	tests := []struct {
		name     string
		bv1      BitVector
		bv2      BitVector
		expected int
	}{
		{
			name:     "nothing in common",
			bv1:      New([]byte{0b11000000}, 2),
			bv2:      New([]byte{0b00000000}, 2),
			expected: 0,
		},
		{
			name:     "a byte and a byte",
			bv1:      New([]byte{0b11110001}, 8),
			bv2:      New([]byte{0b11110010}, 8),
			expected: 6,
		},
		{
			name:     "some bits and a byte",
			bv1:      New([]byte{0b11110000}, 4),
			bv2:      New([]byte{0b11100111}, 8),
			expected: 3,
		},
		{
			name:     "some bits and a byte",
			bv1:      New([]byte{0b11110000}, 4),
			bv2:      New([]byte{0b11110111}, 8),
			expected: 4,
		},
		{
			name:     "differ before a byte boundary",
			bv1:      New([]byte{0b00111000, 0b10111001, 0b01000010}, 24),
			bv2:      New([]byte{0b00111000, 0b10111000, 0b01000010}, 24),
			expected: 15,
		},
		{
			name:     "differ on a byte boundary",
			bv1:      New([]byte{0b00111000, 0b10111001, 0b11000010}, 24),
			bv2:      New([]byte{0b00111000, 0b10111001, 0b01000010}, 24),
			expected: 16,
		},
		{
			name:     "differ after a byte boundary",
			bv1:      New([]byte{0b00111000, 0b10111001, 0b01000010}, 24),
			bv2:      New([]byte{0b00111000, 0b10111001, 0b00100010}, 24),
			expected: 17,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(
				t,
				test.expected,
				test.bv1.CommonPrefixLen(test.bv2),
			)
		})
	}
}
