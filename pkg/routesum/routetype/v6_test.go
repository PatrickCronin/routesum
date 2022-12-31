package routetype

import (
	"testing"

	"github.com/PatrickCronin/routesum/pkg/routesum/uint128"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseV6String(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedV6    *V6
		expectedError string
	}{
		{
			name:  "valid IPv6 address",
			input: "2001:db8::",
			expectedV6: &V6{
				ip:   uint128.New(2306139568115548160, 0),
				bits: 128,
			},
		},
		{
			name:  "valid: IPv4-embedded IPv6 address",
			input: "::ffff:192.0.2.0",
			expectedV6: &V6{
				ip:   uint128.New(0, 281473902969344),
				bits: 128,
			},
		},
		{
			name:          "invalid: IPv4 address",
			input:         "192.0.2.0",
			expectedError: "`192.0.2.0` is neither an IPv6 nor an IPv4-embedded IPv6 address",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotV6, err := ParseV6String(test.input)

			if test.expectedV6 != nil {
				assert.Equal(t, test.expectedV6, gotV6)
			} else {
				assert.Nil(t, gotV6)
			}

			if test.expectedError != "" {
				assert.EqualError(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestV6String(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "v6 IP",
			input:    "2001:db8::1",
			expected: "2001:db8::1",
		},
		{
			name:     "v6 single-host network",
			input:    "2001:db8::2/128",
			expected: "2001:db8::2",
		},
		{
			name:     "v6 multi-host network",
			input:    "2001:db8::16/126",
			expected: "2001:db8::16/126",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := MustParseV6String(test.input).String()
			require.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestV6Contains(t *testing.T) {
	tests := []struct {
		name           string
		a, b           *V6
		expectContains bool
	}{
		{
			name:           "a contains b",
			a:              MustParseV6String("2001:db8::0/125"),
			b:              MustParseV6String("2001:db8::4/128"),
			expectContains: true,
		},
		{
			name:           "a equals b",
			a:              MustParseV6String("2001:db8::0/125"),
			b:              MustParseV6String("2001:db8::0/125"),
			expectContains: true,
		},
		{
			name:           "a contains b, both share base IP",
			a:              MustParseV6String("2001:db8::/120"),
			b:              MustParseV6String("2001:db8::/121"),
			expectContains: true,
		},
		{
			name:           "a does not contain b, both share base IP",
			a:              MustParseV6String("2001:db8::/121"),
			b:              MustParseV6String("2001:db8::/120"),
			expectContains: false,
		},
		{
			name:           "a does not contain b, both share bit length",
			a:              MustParseV6String("2001:db8::0/126"),
			b:              MustParseV6String("2001:db8::4/126"),
			expectContains: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectContains, test.a.Contains(test.b))
		})
	}
}

func TestV6CommonAncestor(t *testing.T) {
	tests := []struct {
		name                   string
		a, b                   *V6
		expectedCommonAncestor *V6
	}{
		{
			name:                   "if a contains b, common ancestor is the same as a",
			a:                      MustParseV6String("2001:db8::/125"),
			b:                      MustParseV6String("2001:db8::4/128"),
			expectedCommonAncestor: MustParseV6String("2001:db8::/125"),
		},
		{
			name:                   "if b contains a, common ancestor is the same as b",
			a:                      MustParseV6String("2001:db8::4/128"),
			b:                      MustParseV6String("2001:db8::/125"),
			expectedCommonAncestor: MustParseV6String("2001:db8::/125"),
		},
		{
			name:                   "if a and b are the same, so is their common ancestor",
			a:                      MustParseV6String("2001:db8::/125"),
			b:                      MustParseV6String("2001:db8::/125"),
			expectedCommonAncestor: MustParseV6String("2001:db8::/125"),
		},
		{
			name:                   "a and b diverge",
			a:                      MustParseV6String("2001:db8::12/127"),
			b:                      MustParseV6String("2001:db8::2e/128"),
			expectedCommonAncestor: MustParseV6String("2001:db8::/122"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedCommonAncestor, test.a.CommonAncestor(test.b))
		})
	}
}

func TestV6NthBit(t *testing.T) {
	tests := []struct {
		route    *V6
		bitN     uint8
		expected uint8
	}{
		{
			route:    MustParseV6String("2001:db8::"),
			bitN:     1,
			expected: 0,
		},
		{
			route:    MustParseV6String("2001:db8::"),
			bitN:     128,
			expected: 0,
		},
		{
			route:    MustParseV6String("2001:db8::"),
			bitN:     2,
			expected: 0,
		},
		{
			route:    MustParseV6String("2001:db8::"),
			bitN:     3,
			expected: 1,
		},
		{
			route:    MustParseV6String("2001:db8::"),
			bitN:     4,
			expected: 0,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.route.NthBit(test.bitN), "bit %d", test.bitN)
	}
}

func TestV6Size(t *testing.T) {
	r := MustParseV6String("2001:db8::")
	assert.Equal(t, uintptr(24), r.Size())
}
