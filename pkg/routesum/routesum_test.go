package routesum

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStrings(t *testing.T) {
	// Summarization logic is tested in TestSummarize.

	// Invalid IPs throw the expected error
	invalidIPs := []string{
		"192.0.2",
		"::ffff:198.51.100",
		"2001:db8:",
		"not an IP",
	}
	invalidIPErr := regexp.MustCompile(`validate IP:.*was not understood\.`)
	for _, invalidIP := range invalidIPs {
		summarized, err := Strings([]string{invalidIP})
		assert.Nil(t, summarized)
		if assert.Error(t, err) {
			var iiErr *InvalidInputErr
			assert.True(t, errors.As(err, &iiErr), "resulting error can be converted to an InvalidInputErr")
			assert.Equal(t, invalidIP, iiErr.InvalidValue, "expected InvalidValue can be extracted")
			assert.Regexp(t, invalidIPErr, err.Error())
		}
	}

	// Invalid networks throw the expected error
	invalidNets := []string{
		"192.0.2/24",
		"192.0.2.0/33",
		"::ffff:198.51.100/120",
		"::ffff:198.51.100.0/129",
		"2001:db8:/32",
		"2001:db8::/129",
		"not/a/network",
	}
	invalidNetErr := regexp.MustCompile(`validate network:.*was not understood\.`)
	for _, invalidNet := range invalidNets {
		summarized, err := Strings([]string{invalidNet})
		assert.Nil(t, summarized)
		if assert.Error(t, err) {
			var iiErr *InvalidInputErr
			assert.True(t, errors.As(err, &iiErr), "resulting error can be converted to an InvalidInputErr")
			assert.Equal(t, invalidNet, iiErr.InvalidValue, "expected InvalidValue can be extracted")
			assert.Regexp(t, invalidNetErr, err.Error())
		}
	}

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "IPs and networks are returned in the form provided",
			input: []string{
				"192.0.2.0",
				"::ffff:192.0.2.0",
				"2001:db8::",
				"198.51.100.0/24",
				"::ffff:198.51.100.0/120",
				"2001:db8:1::/48",
			},
			expected: []string{
				"192.0.2.0",
				"::ffff:192.0.2.0",
				"2001:db8::",
				"198.51.100.0/24",
				"::ffff:198.51.100.0/120",
				"2001:db8:1::/48",
			},
		},
		{
			name: "documented caveat: IPs and zero-host networks are thought to be the same",
			input: []string{
				"192.0.2.0",
				"::ffff:192.0.2.0",
				"2001:db8::",
				"192.0.2.0/32",
				"::ffff:192.0.2.0/128",
				"2001:db8::/128",
			},
			expected: []string{
				"192.0.2.0",
				"::ffff:192.0.2.0",
				"2001:db8::",
			},
		},
		{
			name: "a simple summarization works as expected",
			input: []string{
				"192.0.2.44",
				"192.0.2.45",
			},
			expected: []string{
				"192.0.2.44/31",
			},
		},
	}
	for _, test := range tests {
		summarized, err := Strings(test.input)
		assert.NoError(t, err, test.name)
		assert.Equal(t, test.expected, summarized, test.name)
	}
}

func TestNetworksAndIPs(t *testing.T) {
	// Summarization logic is tested in TestSummarize.

	// Invalid IPs throw the expected error
	invalidIPs := []net.IP{
		[]byte{192, 0, 2},
		[]byte{192, 0, 2, 0, 0},
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 198, 51, 100},
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 198, 51, 100, 0, 0},
		[]byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	invalidIPErr := regexp.MustCompile(`validate IP:.*was not understood\.`)
	for _, invalidIP := range invalidIPs {
		sumNets, sumIPs, err := NetworksAndIPs([]net.IPNet{}, []net.IP{invalidIP})
		assert.Nil(t, sumNets)
		assert.Nil(t, sumIPs)
		if assert.Error(t, err) {
			var iiErr *InvalidInputErr
			assert.True(t, errors.As(err, &iiErr), "resulting error can be converted to an InvalidInputErr")
			assert.Equal(t, fmt.Sprintf("%#v", invalidIP), iiErr.InvalidValue, "expected InvalidValue can be extracted")
			assert.Regexp(t, invalidIPErr, err.Error())
		}
	}

	// Invalid networks throw the expected error
	invalidNetIPs := []net.IPNet{
		{
			IP:   []byte{192, 0, 2},
			Mask: []byte{255, 255, 0},
		},
		{
			IP:   []byte{192, 0, 2, 0, 0},
			Mask: []byte{255, 255, 255, 255, 0},
		},
		{
			IP:   []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 198, 51, 100},
			Mask: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0},
		},
		{
			IP:   []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 198, 51, 100, 0, 0},
			Mask: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0},
		},
		{
			IP:   []byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Mask: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0},
		},
		{
			IP:   []byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Mask: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0},
		},
	}
	invalidNetIPErr := regexp.MustCompile(`validate network:.*was not understood\.`)
	for _, invalidNetIP := range invalidNetIPs {
		sumNets, sumIPs, err := NetworksAndIPs([]net.IPNet{invalidNetIP}, []net.IP{})
		assert.Nil(t, sumNets)
		assert.Nil(t, sumIPs)
		if assert.Error(t, err) {
			assert.Regexp(t, invalidNetIPErr, err.Error())
		}
	}

	invalidNetMasks := []net.IPNet{
		{
			IP:   []byte{192, 0, 2, 0},
			Mask: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			IP:   []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 198, 51, 100, 0},
			Mask: []byte{255, 255, 255, 0},
		},
		{
			IP:   []byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Mask: []byte{255, 255, 255, 0},
		},
	}
	invalidNetMaskErr := regexp.MustCompile(`validate network:.*was not understood\.`)
	for _, invalidNetIP := range invalidNetMasks {
		sumNets, sumIPs, err := NetworksAndIPs([]net.IPNet{invalidNetIP}, []net.IP{})
		assert.Nil(t, sumNets)
		assert.Nil(t, sumIPs)
		if assert.Error(t, err) {
			assert.Regexp(t, invalidNetMaskErr, err.Error())
		}
	}

	tests := []struct {
		name             string
		inputNetworks    []net.IPNet
		inputIPs         []net.IP
		expectedNetworks []net.IPNet
		expectedIPs      []net.IP
	}{
		{
			name: "IPs and networks are returned in the form provided",
			inputNetworks: []net.IPNet{
				{
					IP:   []byte{198, 51, 100, 0},
					Mask: []byte{255, 255, 255, 0},
				},
				{
					IP:   []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 198, 51, 100, 0},
					Mask: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0},
				},
				{
					IP:   []byte{0x20, 0x01, 0x0d, 0xb8, 0, 0x01, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					Mask: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
			},
			inputIPs: []net.IP{
				[]byte{192, 0, 2, 0},
				[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 192, 0, 2, 0},
				[]byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			},
			expectedNetworks: []net.IPNet{
				{
					IP:   []byte{198, 51, 100, 0},
					Mask: []byte{255, 255, 255, 0},
				},
				{
					IP:   []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 198, 51, 100, 0},
					Mask: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0},
				},
				{
					IP:   []byte{0x20, 0x01, 0x0d, 0xb8, 0, 0x01, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					Mask: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
			},
			expectedIPs: []net.IP{
				[]byte{192, 0, 2, 0},
				[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 192, 0, 2, 0},
				[]byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			},
		},
		{
			name: "documented caveat: IPs and zero-host networks are thought to be the same",
			inputNetworks: []net.IPNet{
				{
					IP:   []byte{192, 0, 2, 0},
					Mask: []byte{255, 255, 255, 255},
				},
				{
					IP:   []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 192, 0, 2, 0},
					Mask: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				{
					IP:   []byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					Mask: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
			},
			inputIPs: []net.IP{
				[]byte{192, 0, 2, 0},
				[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 192, 0, 2, 0},
				[]byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			},
			expectedNetworks: []net.IPNet{},
			expectedIPs: []net.IP{
				[]byte{192, 0, 2, 0},
				[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 192, 0, 2, 0},
				[]byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			},
		},
		{
			name:          "a simple summarization works as expected",
			inputNetworks: []net.IPNet{},
			inputIPs: []net.IP{
				[]byte{192, 0, 2, 44},
				[]byte{192, 0, 2, 45},
			},
			expectedNetworks: []net.IPNet{
				{
					IP:   []byte{192, 0, 2, 44},
					Mask: []byte{255, 255, 255, 254},
				},
			},
			expectedIPs: []net.IP{},
		},
	}
	for _, test := range tests {
		sumNets, sumIPs, err := NetworksAndIPs(test.inputNetworks, test.inputIPs)
		assert.NoError(t, err, test.name)
		assert.Equal(t, test.expectedNetworks, sumNets, test.name)
		assert.Equal(t, test.expectedIPs, sumIPs, test.name)
	}
}

func TestSummarize(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "duplicate IPs and networks are removed",
			input: []string{
				"192.0.2.0",
				"192.0.2.0",
				"::ffff:192.0.2.0",
				"::ffff:192.0.2.0",
				"2001:db8::",
				"2001:db8::",
				"198.51.100.0/31",
				"198.51.100.0/31",
				"::ffff:198.51.100.0/127",
				"::ffff:198.51.100.0/127",
				"2001:db8::1:0/127",
				"2001:db8::1:0/127",
			},
			expected: []string{
				"192.0.2.0",
				"::ffff:192.0.2.0",
				"2001:db8::",
				"198.51.100.0/31",
				"::ffff:198.51.100.0/127",
				"2001:db8::1:0/127",
			},
		},
		{
			name: "covered networks are removed",
			input: []string{
				"192.0.2.16",
				"192.0.2.0/26",
				"::ffff:192.0.2.16",
				"::ffff:192.0.2.0/120",
				"2001:db8::10",
				"2001:db8::/120",
			},
			expected: []string{
				"192.0.2.0/26",
				"::ffff:192.0.2.0/120",
				"2001:db8::/120",
			},
		},
		{
			name: "IPs to networks: one step",
			input: []string{
				"192.0.2.0",
				"192.0.2.1",
				"::ffff:192.0.2.0",
				"::ffff:192.0.2.1",
				"2001:db8::",
				"2001:db8::1",
			},
			expected: []string{
				"192.0.2.0/31",
				"::ffff:192.0.2.0/127",
				"2001:db8::/127",
			},
		},
		{
			name: "IPs to networks: two steps",
			input: []string{
				"192.0.2.0",
				"192.0.2.1",
				"192.0.2.2",
				"192.0.2.3",
				"::ffff:192.0.2.0",
				"::ffff:192.0.2.1",
				"::ffff:192.0.2.2",
				"::ffff:192.0.2.3",
				"2001:db8::",
				"2001:db8::1",
				"2001:db8::2",
				"2001:db8::3",
			},
			expected: []string{
				"192.0.2.0/30",
				"::ffff:192.0.2.0/126",
				"2001:db8::/126",
			},
		},
		{
			name: "cascading summarization",
			input: []string{
				"192.0.2.0",
				"192.0.2.1",
				"192.0.2.2/31",
				"192.0.2.4/30",
				"192.0.2.8/29",
				"192.0.2.16/28",
				"192.0.2.32/27",
				"192.0.2.64/26",
				"192.0.2.128/25",
				"::ffff:192.0.2.0",
				"::ffff:192.0.2.1",
				"::ffff:192.0.2.2/127",
				"::ffff:192.0.2.4/126",
				"::ffff:192.0.2.8/125",
				"::ffff:192.0.2.16/124",
				"::ffff:192.0.2.32/123",
				"::ffff:192.0.2.64/122",
				"::ffff:192.0.2.128/121",
				"2001:db8::",
				"2001:db8::1",
				"2001:db8::2/127",
				"2001:db8::4/126",
				"2001:db8::8/125",
				"2001:db8::10/124",
				"2001:db8::20/123",
				"2001:db8::40/122",
				"2001:db8::80/121",
				"2001:db8::100/120",
				"2001:db8::200/119",
				"2001:db8::400/118",
				"2001:db8::800/117",
				"2001:db8::1000/116",
				"2001:db8::2000/115",
				"2001:db8::4000/114",
				"2001:db8::8000/113",
				"2001:db8::1:0/112",
				"2001:db8::2:0/111",
				"2001:db8::4:0/110",
				"2001:db8::8:0/109",
				"2001:db8::10:0/108",
				"2001:db8::20:0/107",
				"2001:db8::40:0/106",
				"2001:db8::80:0/105",
				"2001:db8::100:0/104",
				"2001:db8::200:0/103",
				"2001:db8::400:0/102",
				"2001:db8::800:0/101",
				"2001:db8::1000:0/100",
				"2001:db8::2000:0/99",
				"2001:db8::4000:0/98",
				"2001:db8::8000:0/97",
				"2001:db8::1:0:0/96",
				"2001:db8::2:0:0/95",
				"2001:db8::4:0:0/94",
				"2001:db8::8:0:0/93",
				"2001:db8::10:0:0/92",
				"2001:db8::20:0:0/91",
				"2001:db8::40:0:0/90",
				"2001:db8::80:0:0/89",
				"2001:db8::100:0:0/88",
				"2001:db8::200:0:0/87",
				"2001:db8::400:0:0/86",
				"2001:db8::800:0:0/85",
				"2001:db8::1000:0:0/84",
				"2001:db8::2000:0:0/83",
				"2001:db8::4000:0:0/82",
				"2001:db8::8000:0:0/81",
				"2001:db8:0:0:1::/80",
				"2001:db8:0:0:2::/79",
				"2001:db8:0:0:4::/78",
				"2001:db8:0:0:8::/77",
				"2001:db8:0:0:10::/76",
				"2001:db8:0:0:20::/75",
				"2001:db8:0:0:40::/74",
				"2001:db8:0:0:80::/73",
				"2001:db8:0:0:100::/72",
				"2001:db8:0:0:200::/71",
				"2001:db8:0:0:400::/70",
				"2001:db8:0:0:800::/69",
				"2001:db8:0:0:1000::/68",
				"2001:db8:0:0:2000::/67",
				"2001:db8:0:0:4000::/66",
				"2001:db8:0:0:8000::/65",
				"2001:db8:0:1::/64",
				"2001:db8:0:2::/63",
				"2001:db8:0:4::/62",
				"2001:db8:0:8::/61",
				"2001:db8:0:10::/60",
				"2001:db8:0:20::/59",
				"2001:db8:0:40::/58",
				"2001:db8:0:80::/57",
				"2001:db8:0:100::/56",
				"2001:db8:0:200::/55",
				"2001:db8:0:400::/54",
				"2001:db8:0:800::/53",
				"2001:db8:0:1000::/52",
				"2001:db8:0:2000::/51",
				"2001:db8:0:4000::/50",
				"2001:db8:0:8000::/49",
				"2001:db8:1::/48",
				"2001:db8:2::/47",
				"2001:db8:4::/46",
				"2001:db8:8::/45",
				"2001:db8:10::/44",
				"2001:db8:20::/43",
				"2001:db8:40::/42",
				"2001:db8:80::/41",
				"2001:db8:100::/40",
				"2001:db8:200::/39",
				"2001:db8:400::/38",
				"2001:db8:800::/37",
				"2001:db8:1000::/36",
				"2001:db8:2000::/35",
				"2001:db8:4000::/34",
				"2001:db8:8000::/33",
			},
			expected: []string{
				"192.0.2.0/24",
				"::ffff:192.0.2.0/120",
				"2001:db8::/32",
			},
		},
		{
			name: "unsummarized IPs and networks are returned",
			input: []string{
				"192.0.2.1",
				"192.0.2.2",
				"192.0.2.3",
				"192.0.2.4",
				"::ffff:192.0.2.1",
				"::ffff:192.0.2.2",
				"::ffff:192.0.2.3",
				"::ffff:192.0.2.4",
				"2001:db8::1",
				"2001:db8::2",
				"2001:db8::3",
				"2001:db8::4",
				"198.51.100.2/31",
				"198.51.100.4/31",
				"198.51.100.6/31",
				"198.51.100.8/31",
				"::ffff:198.51.100.2/127",
				"::ffff:198.51.100.4/127",
				"::ffff:198.51.100.6/127",
				"::ffff:198.51.100.8/127",
				"2001:db8::1:2/127",
				"2001:db8::1:4/127",
				"2001:db8::1:6/127",
				"2001:db8::1:8/127",
			},
			expected: []string{
				"192.0.2.1",
				"192.0.2.4",
				"::ffff:192.0.2.1",
				"::ffff:192.0.2.4",
				"2001:db8::1",
				"2001:db8::4",
				"192.0.2.2/31",
				"198.51.100.2/31",
				"198.51.100.8/31",
				"198.51.100.4/30",
				"::ffff:192.0.2.2/127",
				"::ffff:198.51.100.2/127",
				"::ffff:198.51.100.8/127",
				"2001:db8::2/127",
				"2001:db8::1:2/127",
				"2001:db8::1:8/127",
				"::ffff:198.51.100.4/126",
				"2001:db8::1:4/126",
			},
		},
		{
			name: "duplicates resulting from summarization are removed",
			input: []string{
				"192.0.2.0",
				"192.0.2.1",
				"192.0.2.0/31",
				"::ffff:192.0.2.0",
				"::ffff:192.0.2.1",
				"::ffff:192.0.2.0/127",
				"2001:db8::",
				"2001:db8::1",
				"2001:db8::/127",
			},
			expected: []string{
				"192.0.2.0/31",
				"::ffff:192.0.2.0/127",
				"2001:db8::/127",
			},
		},
		{
			name: "documented caveat: IPv4-embedded IPv6 address treated as IPv6 version",
			input: []string{
				"::ffff:192.0.2.0",
			},
			expected: []string{
				"::ffff:192.0.2.0",
			},
		},
		{
			name: "documented caveat: IPv4-embedded IPv6 addresses are treated differently",
			input: []string{
				"192.0.2.0",
				"::ffff:192.0.2.0",
			},
			expected: []string{
				"192.0.2.0",
				"::ffff:192.0.2.0",
			},
		},
		{
			name: "documented caveat: zero-host networks are treated as IPv4 addresses",
			input: []string{
				"192.0.2.0/32",
				"::ffff:192.0.2.0/128",
				"2001:db8::/128",
			},
			expected: []string{
				"192.0.2.0",
				"::ffff:192.0.2.0",
				"2001:db8::",
			},
		},
		{
			name: "documented caveat: zero-host networks are removed as duplicates if IP counterparts are already present",
			input: []string{
				"192.0.2.0",
				"192.0.2.0/32",
				"::ffff:192.0.2.0",
				"::ffff:192.0.2.0/128",
				"2001:db8::",
				"2001:db8::/128",
			},
			expected: []string{
				"192.0.2.0",
				"::ffff:192.0.2.0",
				"2001:db8::",
			},
		},
	}

	for _, test := range tests {
		strs, err := Strings(test.input)
		require.NoError(t, err, test.name)
		assert.Equal(t, test.expected, strs, test.name)
	}
}

func TestRemoveContainedNetworks(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "networks contained by other networks are removed",
			input: []string{
				"192.0.2.27/32",
				"192.0.2.16/28",
				"192.0.2.8/27",
				"::ffff:192.0.2.27/128",
				"::ffff:192.0.2.16/124",
				"::ffff:192.0.2.8/123",
				"2001:db8::1b/128",
				"2001:db8::10/124",
				"2001:db8::8/123",
			},
			expected: []string{
				"2001:db8::8/123",
				"::ffff:192.0.2.8/123",
				"192.0.2.8/27",
			},
		},
		{
			name: "list with no contained networks is not modified",
			input: []string{
				"192.0.2.16/32",
				"198.51.100.0/29",
				"203.0.113.0/30",
				"2001:db8::/120",
			},
			expected: []string{
				"2001:db8::/120",
				"198.51.100.0/29",
				"203.0.113.0/30",
				"192.0.2.16/32",
			},
		},
	}

	for _, test := range tests {
		var inputSRNets []safeRepNet
		for _, in := range test.input {
			srNet, err := newSafeRepNetFromString(in)
			require.NoError(t, err)
			inputSRNets = append(inputSRNets, *srNet)
		}

		var expectedSRNets []safeRepNet
		for _, exp := range test.expected {
			srNet, err := newSafeRepNetFromString(exp)
			require.NoError(t, err)
			expectedSRNets = append(expectedSRNets, *srNet)
		}

		uncontainedSRNets := removeContainedNetworks(inputSRNets)
		assert.Equal(t, expectedSRNets, uncontainedSRNets, test.name)
	}
}

func TestTrySumNets(t *testing.T) {
	tests := []struct {
		name     string
		nets     [2]string
		expected string
	}{
		{
			name: "IPv4 and IPv4-embedded IPv6 networks don't summarize",
			nets: [2]string{
				"192.0.2.0/32",
				"::ffff:192.0.2.1/128",
			},
			expected: "",
		},
		{
			name: "IPv4-embedded IPv6 and IPv6 networks do summarize",
			nets: [2]string{
				"::ffff:192.0.2.0/128",
				"::ffff:c000:201/128", // ::ffff:192.0.2.1/128
			},
			expected: "::ffff:c000:200/127",
		},
		{
			name: "IPv4: duplicate networks don't summarize",
			nets: [2]string{
				"192.0.2.0/32",
				"192.0.2.0/32",
			},
			expected: "",
		},
		{
			name: "IPv4-embedded IPv6: duplicate networks don't summarize",
			nets: [2]string{
				"::ffff:192.0.2.0/128",
				"::ffff:192.0.2.0/128",
			},
			expected: "",
		},
		{
			name: "IPv6: duplicate networks don't summarize",
			nets: [2]string{
				"2001:db8::/128",
				"2001:db8::/128",
			},
			expected: "",
		},
		{
			name: "IPv4: consecutive but misaligned networks don't summarize",
			nets: [2]string{
				"192.0.2.1/32",
				"192.0.2.2/32",
			},
			expected: "",
		},
		{
			name: "IPv4-embedded IPv6: consecutive but misaligned networks don't summarize",
			nets: [2]string{
				"::ffff:192.0.2.1/128",
				"::ffff:192.0.2.2/128",
			},
			expected: "",
		},
		{
			name: "IPv6: consecutive but misaligned networks don't summarize",
			nets: [2]string{
				"2001:db8::1/128",
				"2001:db8::2/128",
			},
			expected: "",
		},
		{
			name: "IPv4: consecutive and aligned networks do summarize",
			nets: [2]string{
				"192.0.2.0/32",
				"192.0.2.1/32",
			},
			expected: "192.0.2.0/31",
		},
		{
			name: "IPv4-embedded IPv6: consecutive and aligned networks do summarize",
			nets: [2]string{
				"::ffff:192.0.2.0/128",
				"::ffff:192.0.2.1/128",
			},
			expected: "::ffff:192.0.2.0/127",
		},
		{
			name: "IPv6: consecutive and aligned networks do summarize",
			nets: [2]string{
				"2001:db8::0/128",
				"2001:db8::1/128",
			},
			expected: "2001:db8::0/127",
		},
		{
			name: "IPv4: order of networks doesn't interfere with summarization",
			nets: [2]string{
				"192.0.2.1/32",
				"192.0.2.0/32",
			},
			expected: "192.0.2.0/31",
		},
		{
			name: "IPv4-embedded IPv6: order of networks doesn't interfere with summarization",
			nets: [2]string{
				"::ffff:192.0.2.1/128",
				"::ffff:192.0.2.0/128",
			},
			expected: "::ffff:192.0.2.0/127",
		},
		{
			name: "IPv6: order of networks doesn't interfere with summarization",
			nets: [2]string{
				"2001:db8::1/128",
				"2001:db8::0/128",
			},
			expected: "2001:db8::0/127",
		},
	}

	for _, test := range tests {
		a, err := newSafeRepNetFromString(test.nets[0])
		require.NoError(t, err)
		b, err := newSafeRepNetFromString(test.nets[1])
		require.NoError(t, err)

		var expected *safeRepNet
		if test.expected != "" {
			exp, err := newSafeRepNetFromString(test.expected)
			require.NoError(t, err)
			expected = exp
		}

		sumNet := trySumNets(*a, *b)
		assert.Equal(t, expected, sumNet, test.name)
	}
}
