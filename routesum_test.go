package routesum

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
	Documentation Networks:
	* 192.0.2.0/24
	* 198.51.100.0/24
	* 203.0.113.0/24
	* 2001:db8::/32
*/
func TestStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "simple IPv4 list",
			input: []string{
				"1.1.1.0",
				"1.1.1.5",
				"1.1.1.4",
				"1.1.1.2",
				"1.1.1.7",
				"207.49.18.0/24",
				"1.1.1.1",
				"1.1.1.5",
				"1.1.1.3",
				"1.1.1.6",
			},
			expected: []string{
				"1.1.1.0/29",
				"207.49.18.0/24",
			},
		},
		{
			name: "summarize: one full /30",
			input: []string{
				"1.1.1.0",
				"1.1.1.1",
				"1.1.1.2",
				"1.1.1.3",
			},
			expected: []string{
				"1.1.1.0/30",
			},
		},
		{
			name: "summarize: 2 IPs, 1 full /31",
			input: []string{
				"1.1.1.1",
				"1.1.1.2",
				"1.1.1.3",
				"1.1.1.4",
			},
			expected: []string{
				"1.1.1.1",
				"1.1.1.4",
				"1.1.1.2/31",
			},
		},
		{
			name: "summarize: two full /31",
			input: []string{
				"1.1.1.2",
				"1.1.1.3",
				"1.1.1.4",
				"1.1.1.5",
			},
			expected: []string{
				"1.1.1.2/31",
				"1.1.1.4/31",
			},
		},
		{
			name: "summarize: consecutive IPs in different /31s",
			input: []string{
				"1.1.1.1",
				"1.1.1.2",
			},
			expected: []string{
				"1.1.1.1",
				"1.1.1.2",
			},
		},
		{
			name: "summarize: consecutive IPs in different /24s",
			input: []string{
				"1.1.1.255",
				"1.1.2.0",
			},
			expected: []string{
				"1.1.1.255",
				"1.1.2.0",
			},
		},
		{
			name: "summarize: networks combine",
			input: []string{
				"2.2.2.8",
				"2.2.2.9",
				"2.2.2.10",
				"2.2.2.11",
				"2.2.2.12",
				"2.2.2.13",
				"2.2.2.14",
				"2.2.2.15",
				"2.2.2.16",
				"2.2.2.17",
				"2.2.2.18",
			},
			expected: []string{
				"2.2.2.18",
				"2.2.2.16/31",
				"2.2.2.8/29",
			},
		},
		{
			name: "duplicate IPs are removed",
			input: []string{
				"192.0.2.0",
				"192.0.2.0",
			},
			expected: []string{
				"192.0.2.0",
			},
		},
		{
			name: "duplicate networks are removed",
			input: []string{
				"192.0.2.0/24",
				"192.0.2.0/24",
			},
			expected: []string{
				"192.0.2.0/24",
			},
		},
		{
			name: "duplicates resulting from summarization are removed",
			input: []string{
				"192.0.2.0",
				"192.0.2.1",
				"192.0.2.0/31",
			},
			expected: []string{
				"192.0.2.0/31",
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
			},
			expected: []string{
				"192.0.2.0",
			},
		},
		{
			name: "documented caveat: zero-host networks are removed as duplicates if IP counterparts are already present",
			input: []string{
				"192.0.2.0",
				"192.0.2.0/32",
			},
			expected: []string{
				"192.0.2.0",
			},
		},
	}

	for _, test := range tests {
		strs, err := Strings(test.input)
		require.NoError(t, err, test.name)
		assert.Equal(t, test.expected, strs, test.name)
	}

}

// func TestTrySumHosts(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		ips      [2]string
// 		expected *net.IPNet
// 	}{
// 		{
// 			name: "different lengths",
// 			ips: [2]string{
// 				"1.1.1.1",
// 				"f2d4:126a:9eca:d7fc:6e1b:6846:5403:f201",
// 			},
// 			expected: nil,
// 		},
// 		{
// 			name: "same IPs",
// 			ips: [2]string{
// 				"1.1.1.0",
// 				"1.1.1.0",
// 			},
// 			expected: nil,
// 		},
// 		{
// 			name: "not in same /31",
// 			ips: [2]string{
// 				"1.1.1.1",
// 				"1.1.1.2",
// 			},
// 			expected: nil,
// 		},
// 		{
// 			name: "simple summarization",
// 			ips: [2]string{
// 				"1.1.1.0",
// 				"1.1.1.1",
// 			},
// 			expected: &net.IPNet{
// 				IP:   canonicalIP(net.ParseIP("1.1.1.0")),
// 				Mask: net.CIDRMask(31, 32),
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		a := canonicalIP(net.ParseIP(test.ips[0]))
// 		b := canonicalIP(net.ParseIP(test.ips[1]))
// 		sumNet := trySumHosts(a, b)
// 		assert.Equal(t, test.expected, sumNet, test.name)
// 	}
// }

// func TestSumRoutes(t *testing.T) {
// 	tests := []struct {
// 		name             string
// 		ips              []string
// 		expectedHosts    []string
// 		expectedNetworks []string
// 	}{
// 		{
// 			name: "consecutive IPs, 2 full /31s",
// 			ips: []string{
// 				"1.1.1.2",
// 				"1.1.1.3",
// 				"1.1.1.4",
// 				"1.1.1.5",
// 			},
// 			expectedHosts: []string{},
// 			expectedNetworks: []string{
// 				"1.1.1.2/31",
// 				"1.1.1.4/31",
// 			},
// 		},
// 		{
// 			name: "consecutive IPs, 1 full /31",
// 			ips: []string{
// 				"1.1.1.1",
// 				"1.1.1.2",
// 				"1.1.1.3",
// 				"1.1.1.4",
// 			},
// 			expectedHosts: []string{
// 				"1.1.1.1",
// 				"1.1.1.4",
// 			},
// 			expectedNetworks: []string{
// 				"1.1.1.2/31",
// 			},
// 		},
// 		{
// 			name: "consecutive IPs in different /31s",
// 			ips: []string{
// 				"1.1.1.1",
// 				"1.1.1.2",
// 			},
// 			expectedHosts: []string{
// 				"1.1.1.1",
// 				"1.1.1.2",
// 			},
// 			expectedNetworks: []string{},
// 		},
// 		{
// 			name: "consecutive IPs in different /24s",
// 			ips: []string{
// 				"1.1.1.255",
// 				"1.1.2.0",
// 			},
// 			expectedHosts: []string{
// 				"1.1.1.255",
// 				"1.1.2.0",
// 			},
// 			expectedNetworks: []string{},
// 		},
// 		{
// 			name: "consecutive IPs create consecutive /31s",
// 			ips: []string{
// 				"1.1.1.0",
// 				"1.1.1.1",
// 				"1.1.1.2",
// 				"1.1.1.3",
// 			},
// 			expectedHosts:    []string{},
// 			expectedNetworks: []string{"1.1.1.0/30"},
// 		},
// 		{
// 			name: "networks combine",
// 			ips: []string{
// 				"2.2.2.8",
// 				"2.2.2.9",
// 				"2.2.2.10",
// 				"2.2.2.11",
// 				"2.2.2.12",
// 				"2.2.2.13",
// 				"2.2.2.14",
// 				"2.2.2.15",
// 				"2.2.2.16",
// 				"2.2.2.17",
// 				"2.2.2.18",
// 			},
// 			expectedHosts: []string{
// 				"2.2.2.18",
// 			},
// 			expectedNetworks: []string{
// 				"2.2.2.16/31",
// 				"2.2.2.8/29",
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		ips := make([]net.IP, len(test.ips))
// 		for i, s := range test.ips {
// 			ips[i] = canonicalIP(net.ParseIP(s))
// 		}

// 		networks, hosts := SumRoutes(ips, []net.IPNet{})

// 		assert.Equal(t, test.expectedHosts, stringifyHosts(hosts), "hosts: "+test.name)
// 		assert.Equal(t, test.expectedNetworks, stringifyNetworks(networks), "networks: "+test.name)
// 	}
// }

// func stringifyNetworks(networks []net.IPNet) []string {
// 	stringified := make([]string, len(networks))
// 	for i, network := range networks {
// 		stringified[i] = network.String()
// 	}
// 	return stringified
// }

// func stringifyHosts(hosts []net.IP) []string {
// 	stringified := make([]string, len(hosts))
// 	for i, host := range hosts {
// 		stringified[i] = host.String()
// 	}
// 	return stringified
// }

/*
func TestSummarizing(t *testing.T)
* lots of scenarios to test here( /32 + /32 = /31, if boundaries line up!)
* duplicate entries are removed
* contained networks are removed
* zero-host networks are sorted and summarized the same as their IP counterparts
* zero-host network + IP counterpart are truncated as duplicates (try ipv4, ipv4-embedded ipv6, and ipv6)

func TestTrySumNetworks(t *testing.T)
* ?
func TestRemoveContainedNetworks(t *testing.T)
* ?
func TestStrings(t *testing.T)
* ipv4 addresses are sorted and summarized as such
* ipv4-embedded ipv6 addresses are sorted and summarized as ipv6 addresses
* ipv6 addresses are sorted and summarized as such
* ipv4-embedded ipv6 addresses remain separate from their ipv4 counterparts
* works for simple example?
func TestNetworksAndIPs(t *testing.T)
* ipv4 addresses are sorted and summarized as such
* ipv4-embedded ipv6 addresses are sorted and summarized as ipv6 addresses
* ipv6 addresses are sorted and summarized as such
* ipv4-embedded ipv6 addresses remain separate from their ipv4 counterparts
* works for simple example?
*/
