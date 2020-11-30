package routesum

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeRepIPFromString(t *testing.T) {
	invalidIPStrs := []string{
		"192.0.2",
		"192.0.2.0.0",
		"192.0.2::",
		"2001:db8:::",
		"::1::2",
		"::ffff::198.51.100.39",
	}
	for _, s := range invalidIPStrs {
		_, err := newSafeRepIPFromString(s)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "interpret IP")
		}
	}

	validIPTests := []struct {
		name  string
		input string
	}{
		{
			name:  "IPv4 IP",
			input: "192.0.2.39",
		},
		{
			name:  "IPv4-embedded IPv6 IP",
			input: "::ffff:198.51.100.39",
		},
		{
			name:  "IPv6 IP",
			input: "2001:db8::4",
		},
	}
	for _, test := range validIPTests {
		t.Run(test.name, func(t *testing.T) {
			srIP, err := newSafeRepIPFromString(test.input) // nolint: scopelint
			assert.NoError(t, err, "construct safeRepIP from string")
			assert.Equal(t, test.input, srIP.String(), "safeRepIP stringifies as expected") // nolint: scopelint
		})
	}
}

func TestSafeRepIPFromNetIP(t *testing.T) {
	// Try an invalid 5-byte net.IP variable
	_, err := newSafeRepIPFromNetIP(net.IP([]byte{0xc0, 0x0, 0x2, 0x0, 0x0}))
	if assert.Error(t, err) {
		assert.Contains(
			t,
			err.Error(),
			"interpret net.IP",
		)
	}

	validIPTests := []struct {
		name     string
		input    net.IP
		expected string
	}{
		{
			name:     "4-byte net.IP from IPv4 address",
			input:    net.ParseIP("192.0.2.0").To4(),
			expected: "192.0.2.0",
		},
		{
			name:     "16-byte net.IP from IPv4 address",
			input:    net.ParseIP("192.0.2.0").To16(),
			expected: "::ffff:192.0.2.0",
		},
		{
			name:     "16-byte net.IP from IPv6 address",
			input:    net.ParseIP("2001::db8").To16(),
			expected: "2001::db8",
		},
	}
	for _, test := range validIPTests {
		t.Run(test.name, func(t *testing.T) {
			srIP, err := newSafeRepIPFromNetIP(test.input) // nolint: scopelint
			assert.NoError(t, err, "construct safeRepIP from net.IP")
			assert.Equal(t, test.expected, srIP.String(), "safeRepIP stringifies as expected") // nolint: scopelint
		})
	}
}

func TestSafeRepNetFromString(t *testing.T) {
	invalidNetStrs := []string{
		"192.0.2/29",
		"192.0.2.0.0/29",
		"192.0.2::/29",
		"192.0.2.0/33",
		"2001:db8:::/50",
		"::1::2/50",
		"::ffff::198.51.100.39/50",
		"2001:db8::/129",
	}
	for _, s := range invalidNetStrs {
		_, err := newSafeRepNetFromString(s)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "interpret network")
		}
	}

	validNetTests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "IPv4 network",
			input:    "192.0.2.0/25",
			expected: "192.0.2.0/25",
		},
		{
			name:     "IPv4-embedded IPv6 network",
			input:    "::ffff:198.51.100.32/124",
			expected: "::ffff:198.51.100.32/124",
		},
		{
			name:     "IPv6 network",
			input:    "2001:db8:8000::/33",
			expected: "2001:db8:8000::/33",
		},
		{
			name:     "off-base IPv4 network",
			input:    "192.0.2.27/29",
			expected: "192.0.2.24/29",
		},
		{
			name:     "off-base IPv4-embedded IPv6 network",
			input:    "::ffff:198.51.100.37/124",
			expected: "::ffff:198.51.100.32/124",
		},
		{
			name:     "off-base IPv6 network",
			input:    "2001:db8:7fff::/33",
			expected: "2001:db8::/33",
		},
	}
	for _, test := range validNetTests {
		t.Run(test.name, func(t *testing.T) {
			srNet, err := newSafeRepNetFromString(test.input) // nolint: scopelint
			assert.NoError(t, err, "construct safeRepNet from string")
			assert.Equal(t, test.expected, srNet.String(), "safeRepNet stringifies as expected") // nolint: scopelint
		})
	}
}

func TestSafeRepNetFromNetIPNet(t *testing.T) { // nolint: funlen
	invalidIPNetTests := []struct {
		name     string
		input    net.IPNet
		expected string
	}{
		{
			name: "5-byte IP",
			input: net.IPNet{
				IP:   net.IP([]byte{0xc0, 0x0, 0x2, 0x0, 0x0}),        // 192.0.2.0.0?
				Mask: net.IPMask([]byte{0xff, 0xff, 0xff, 0xff, 0x0}), // 255.255.255.255.0?
			},
			expected: "interpret net.IPNet.IP",
		},
		{
			name: "4-byte IPv4 IP with 16-byte mask",
			input: net.IPNet{
				IP: net.IP([]byte{0xc0, 0x0, 0x2, 0x0}), // 192.0.2.0
				Mask: net.IPMask([]byte{ // ffff:ffff:ffff:ffff:ffff:ffff:ffff:ff00
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0,
				}),
			},
			expected: "network IP and mask are different lengths",
		},
		{
			name: "16-byte IPv6 IP with 4-byte mask",
			input: net.IPNet{
				IP: net.IP([]byte{ // 2001:db8::
					0x20, 0x1, 0xd, 0xb8, 0x0, 0x0, 0x0, 0x0,
					0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
				}),
				Mask: net.IPMask([]byte{0xff, 0xff, 0xff, 0x0}), // 255.255.255.0
			},
			expected: "network IP and mask are different lengths",
		},
		{
			name: "16-byte IPv4-embedded IPv6 IP with 4-byte mask",
			input: net.IPNet{
				IP: net.IP([]byte{ // ::ffff:192.0.2.0
					0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
					0x0, 0x0, 0xff, 0xff, 0xc0, 0x0, 0x2, 0x0,
				}),
				Mask: net.IPMask([]byte{0xff, 0xff, 0xff, 0x0}), // 255.255.255.0
			},
			expected: "",
		},
	}
	for _, test := range invalidIPNetTests {
		t.Run(test.name, func(t *testing.T) {
			_, err := newSafeRepNetFromNetIPNet(test.input) // nolint: scopelint
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), test.expected) // nolint: scopelint
			}
		})
	}

	validIPNetTests := []struct {
		name     string
		input    net.IPNet
		expected string
	}{
		{
			name: "4-byte net.IPNet from IPv4 address",
			input: net.IPNet{
				IP:   net.ParseIP("192.0.2.0").To4(),
				Mask: net.IPMask([]byte{0xff, 0xff, 0xff, 0x0}),
			},
			expected: "192.0.2.0/24",
		},
		{
			name: "16-byte net.IPNet from IPv4 address",
			input: net.IPNet{
				IP: net.ParseIP("192.0.2.0").To16(),
				Mask: net.IPMask([]byte{
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0,
				}),
			},
			expected: "::ffff:192.0.2.0/120",
		},
		{
			name: "16-byte net.IPNet from IPv6 address",
			input: net.IPNet{
				IP: net.ParseIP("2001::db8").To16(),
				Mask: net.IPMask([]byte{
					0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0,
					0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
				}),
			},
			expected: "2001::db8/40",
		},
	}
	for _, test := range validIPNetTests {
		t.Run(test.name, func(t *testing.T) {
			srIP, err := newSafeRepNetFromNetIPNet(test.input) // nolint: scopelint
			assert.NoError(t, err, "construct safeRepNet from net.IPNet")
			assert.Equal(t, test.expected, srIP.String(), "safeRepNet stringifies as expected") // nolint: scopelint
		})
	}
}
