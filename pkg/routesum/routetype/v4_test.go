package routetype

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseV4String(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedV4    *V4
		expectedError string
	}{
		{
			name:  "valid IPv4 address",
			input: "192.0.2.0",
			expectedV4: &V4{
				ip:   uint32(3221225984),
				bits: 32,
			},
		},
		{
			name:          "invalid: IPv4-embedded IPv6 address",
			input:         "::ffff:192.0.2.0",
			expectedError: "`::ffff:192.0.2.0` is not an IPv4 address",
		},
		{
			name:          "invalid: IPv6 address",
			input:         "2001:db8::",
			expectedError: "`2001:db8::` is not an IPv4 address",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotV4, err := ParseV4String(test.input)

			if test.expectedV4 != nil {
				assert.Equal(t, test.expectedV4, gotV4)
			} else {
				assert.Nil(t, gotV4)
			}

			if test.expectedError != "" {
				assert.EqualError(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestV4String(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "v4 IP",
			input:    "192.0.2.15",
			expected: "192.0.2.15",
		},
		{
			name:     "v4 single-host network",
			input:    "192.0.2.1/32",
			expected: "192.0.2.1",
		},
		{
			name:     "v4 multi-host network",
			input:    "192.0.2.16/22",
			expected: "192.0.2.16/22",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := MustParseV4String(test.input).String()
			require.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestV4Contains(t *testing.T) {
	tests := []struct {
		name           string
		a, b           *V4
		expectContains bool
	}{
		{
			name:           "a contains b",
			a:              MustParseV4String("192.0.2.0/29"),
			b:              MustParseV4String("192.0.2.4/32"),
			expectContains: true,
		},
		{
			name:           "a equals b",
			a:              MustParseV4String("192.0.2.0/25"),
			b:              MustParseV4String("192.0.2.0/25"),
			expectContains: true,
		},
		{
			name:           "a contains b, both share base IP",
			a:              MustParseV4String("192.0.2.0/25"),
			b:              MustParseV4String("192.0.2.0/26"),
			expectContains: true,
		},
		{
			name:           "a does not contain b, both share base IP",
			a:              MustParseV4String("192.0.2.0/26"),
			b:              MustParseV4String("192.0.2.0/25"),
			expectContains: false,
		},
		{
			name:           "a does not contain b, both share bit length",
			a:              MustParseV4String("192.0.2.0/30"),
			b:              MustParseV4String("192.0.2.4/30"),
			expectContains: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectContains, test.a.Contains(test.b))
		})
	}
}

func TestV4CommonAncestor(t *testing.T) {
	tests := []struct {
		name                   string
		a, b                   *V4
		expectedCommonAncestor *V4
	}{
		{
			name:                   "if a contains b, common ancestor is the same as a",
			a:                      MustParseV4String("192.0.2.0/29"),
			b:                      MustParseV4String("192.0.2.4/32"),
			expectedCommonAncestor: MustParseV4String("192.0.2.0/29"),
		},
		{
			name:                   "if b contains a, common ancestor is the same as b",
			a:                      MustParseV4String("192.0.2.4/32"),
			b:                      MustParseV4String("192.0.2.0/29"),
			expectedCommonAncestor: MustParseV4String("192.0.2.0/29"),
		},
		{
			name:                   "if a and b are the same, so is their common ancestor",
			a:                      MustParseV4String("192.0.2.0/25"),
			b:                      MustParseV4String("192.0.2.0/25"),
			expectedCommonAncestor: MustParseV4String("192.0.2.0/25"),
		},
		{
			name:                   "a and b diverge",
			a:                      MustParseV4String("192.0.2.18/31"),
			b:                      MustParseV4String("192.0.2.46/32"),
			expectedCommonAncestor: MustParseV4String("192.0.2.0/26"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedCommonAncestor, test.a.CommonAncestor(test.b))
		})
	}
}

func TestV4NthBit(t *testing.T) {
	tests := []struct {
		route    *V4
		bitN     uint8
		expected uint8
	}{
		{
			route:    MustParseV4String("192.0.2.0"),
			bitN:     1,
			expected: 1,
		},
		{
			route:    MustParseV4String("192.0.2.0"),
			bitN:     32,
			expected: 0,
		},
		{
			route:    MustParseV4String("192.0.2.0"),
			bitN:     22,
			expected: 0,
		},
		{
			route:    MustParseV4String("192.0.2.0"),
			bitN:     23,
			expected: 1,
		},
		{
			route:    MustParseV4String("192.0.2.0"),
			bitN:     24,
			expected: 0,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.route.NthBit(test.bitN), "bit %d", test.bitN)
	}
}

func TestV4Size(t *testing.T) {
	r := MustParseV4String("192.0.2.0")
	assert.Equal(t, uintptr(8), r.Size())
}
