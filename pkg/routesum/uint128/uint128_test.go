package uint128

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMask6(t *testing.T) { //nolint: funlen
	tests := []struct {
		input    uint8
		expected *Uint128
	}{
		{
			input: 1,
			expected: &Uint128{
				hi: 0x8000000000000000,
				lo: 0,
			},
		},
		{
			input: 2,
			expected: &Uint128{
				hi: 0xC000000000000000,
				lo: 0,
			},
		},
		{
			input: 3,
			expected: &Uint128{
				hi: 0xE000000000000000,
				lo: 0,
			},
		},
		{
			input: 4,
			expected: &Uint128{
				hi: 0xF000000000000000,
				lo: 0,
			},
		},
		{
			input: 5,
			expected: &Uint128{
				hi: 0xF800000000000000,
				lo: 0,
			},
		},
		{
			input: 32,
			expected: &Uint128{
				hi: 0xFFFFFFFF00000000,
				lo: 0,
			},
		},
		{
			input: 63,
			expected: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFE,
				lo: 0,
			},
		},
		{
			input: 64,
			expected: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0,
			},
		},
		{
			input: 65,
			expected: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0x8000000000000000,
			},
		},
		{
			input: 87,
			expected: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0xFFFFFE0000000000,
			},
		},
		{
			input: 127,
			expected: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0xFFFFFFFFFFFFFFFE,
			},
		},
		{
			input: 128,
			expected: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0xFFFFFFFFFFFFFFFF,
			},
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, Mask6(test.input), "Mask6(%d)", test.input)
	}
}

func TestAnd(t *testing.T) {
	tests := []struct {
		name     string
		input1   *Uint128
		input2   *Uint128
		expected *Uint128
	}{
		{
			name: "ones and ones",
			input1: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0xFFFFFFFFFFFFFFFF,
			},
			input2: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0xFFFFFFFFFFFFFFFF,
			},
			expected: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0xFFFFFFFFFFFFFFFF,
			},
		},
		{
			name: "ones and zeros",
			input1: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0xFFFFFFFFFFFFFFFF,
			},
			input2: &Uint128{
				hi: 0,
				lo: 0,
			},
			expected: &Uint128{
				hi: 0,
				lo: 0,
			},
		},
		{
			name: "bit crisscross",
			input1: &Uint128{
				hi: 0xAAAAAAAAAAAAAAAA,
				lo: 0xAAAAAAAAAAAAAAAA,
			},
			input2: &Uint128{
				hi: 0x5555555555555555,
				lo: 0x5555555555555555,
			},
			expected: &Uint128{
				hi: 0x0000000000000000,
				lo: 0x0000000000000000,
			},
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.input1.And(test.input2), test.name)
	}
}

func TestCommonPrefixLen(t *testing.T) { //nolint: funlen
	tests := []struct {
		name     string
		u1, u2   *Uint128
		expected uint8
	}{
		{
			name: "nothing common",
			u1: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0xFFFFFFFFFFFFFFFF,
			},
			u2: &Uint128{
				hi: 0x0000000000000000,
				lo: 0x0000000000000000,
			},
			expected: 0,
		},
		{
			name: "everything common",
			u1: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0xFFFFFFFFFFFFFFFF,
			},
			u2: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0xFFFFFFFFFFFFFFFF,
			},
			expected: 128,
		},
		{
			name: "diverging in the middle of hi",
			u1: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0xFFFFFFFFFFFFFFFF,
			},
			u2: &Uint128{
				hi: 0xFFFFFFFFFF000000,
				lo: 0x0000000000000000,
			},
			expected: 40,
		},
		{
			name: "diverging in the middle of lo",
			u1: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0xFFFFFFFFFFFFFFFF,
			},
			u2: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0xFFFFFFFFFF000000,
			},
			expected: 104,
		},
		{
			name: "just before the hi/lo split",
			u1: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0xFFFFFFFFFFFFFFFF,
			},
			u2: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFE,
				lo: 0xFFFFFFFFFFFFFFFF,
			},
			expected: 63,
		},
		{
			name: "just after the hi/lo split",
			u1: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0x7FFFFFFFFFFFFFFF,
			},
			u2: &Uint128{
				hi: 0xFFFFFFFFFFFFFFFF,
				lo: 0xFFFFFFFFFFFFFFFF,
			},
			expected: 64,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.u1.CommonPrefixLen(test.u2))
	}
}

func TestNthBit(t *testing.T) {
	tests := []struct {
		name     string
		u        *Uint128
		expected func(uint8) uint8
	}{
		{
			name: "one zero one zero ...",
			u: &Uint128{
				hi: 0xAAAAAAAAAAAAAAAA,
				lo: 0xAAAAAAAAAAAAAAAA,
			},
			expected: func(n uint8) uint8 {
				return n % 2
			},
		},
		{
			name: "zero one zero one ...",
			u: &Uint128{
				hi: 0x5555555555555555,
				lo: 0x5555555555555555,
			},
			expected: func(n uint8) uint8 {
				return (n + 1) % 2
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for n := uint8(1); n <= 128; n++ {
				assert.Equal(t, test.expected(n), test.u.NthBit(n), "bit %d", n)
			}
		})
	}
}
