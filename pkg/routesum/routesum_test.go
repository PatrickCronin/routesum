package routesum

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStrings(t *testing.T) { // nolint: funlen
	// Summarization logic is tested in TestSummarize.

	// Invalid IPs throw the expected error
	invalidIPs := []string{
		"192.0.2",
		"::ffff:198.51.100",
		"2001:db8:",
		"not an IP",
	}
	invalidIPErr := regexp.MustCompile(`ParseIP`)
	for _, invalidIP := range invalidIPs {
		t.Run(invalidIP, func(t *testing.T) {
			rs := NewRouteSum()
			err := rs.InsertFromString(invalidIP)
			if assert.Error(t, err) {
				assert.Regexp(t, invalidIPErr, err.Error())
			}
			assert.Equal(t, []string{}, rs.SummaryStrings(), "nothing was added")
		})
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
	invalidNetErr := regexp.MustCompile(`ParseIPPrefix`)
	for _, invalidNet := range invalidNets {
		t.Run(invalidNet, func(t *testing.T) {
			rs := NewRouteSum()
			err := rs.InsertFromString(invalidNet)
			if assert.Error(t, err) {
				assert.Regexp(t, invalidNetErr, err.Error())
			}
			assert.Equal(t, []string{}, rs.SummaryStrings(), "nothing was added")
		})
	}

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "IPv4, IPv6-embedded IPv4 and IPv6 formats are preserved",
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
				"198.51.100.0/24",
				"::ffff:c000:200",
				"::ffff:c633:6400/120",
				"2001:db8::",
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
				"::ffff:c000:200",
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
		t.Run(test.name, func(t *testing.T) {
			rs := NewRouteSum()
			for _, str := range test.input {
				err := rs.InsertFromString(str)
				require.NoError(t, err)
			}
			assert.Equal(t, test.expected, rs.SummaryStrings(), "summarized as expected")
		})
	}
}

func TestSummarize(t *testing.T) { // nolint: funlen
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
				"198.51.100.0/31",
				"::ffff:c000:200",
				"::ffff:c633:6400/127",
				"2001:db8::",
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
				"::ffff:c000:200/120",
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
				"::ffff:c000:200/127",
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
				"::ffff:c000:200/126",
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
				"::ffff:c000:200/120",
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
				"192.0.2.2/31",
				"192.0.2.4",
				"198.51.100.2/31",
				"198.51.100.4/30",
				"198.51.100.8/31",
				"::ffff:c000:201",
				"::ffff:c000:202/127",
				"::ffff:c000:204",
				"::ffff:c633:6402/127",
				"::ffff:c633:6404/126",
				"::ffff:c633:6408/127",
				"2001:db8::1",
				"2001:db8::2/127",
				"2001:db8::4",
				"2001:db8::1:2/127",
				"2001:db8::1:4/126",
				"2001:db8::1:8/127",
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
				"::ffff:c000:200/127",
				"2001:db8::/127",
			},
		},
		{
			name: "documented caveat: IPv4-embedded IPv6 address treated as IPv6 version",
			input: []string{
				"::ffff:192.0.2.0",
			},
			expected: []string{
				"::ffff:c000:200",
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
				"::ffff:c000:200",
			},
		},
		{
			name: "documented caveat: zero-host networks are treated as IP addresses",
			input: []string{
				"192.0.2.0/32",
				"::ffff:192.0.2.0/128",
				"2001:db8::/128",
			},
			expected: []string{
				"192.0.2.0",
				"::ffff:c000:200",
				"2001:db8::",
			},
		},
		{
			name: "documented caveat: zero-host networks are removed if IP counterparts are already present",
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
				"::ffff:c000:200",
				"2001:db8::",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rs := NewRouteSum()
			for _, str := range test.input {
				err := rs.InsertFromString(str)
				require.NoError(t, err)
			}
			assert.Equal(t, test.expected, rs.SummaryStrings(), "got expected summary")
		})
	}
}
