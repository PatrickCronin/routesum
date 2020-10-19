package routesum;

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanonicalizeHosts (t *testing.T) {
	tests := []struct{
		name string
		host net.IP
		expected net.IP
	}{
		{
			name: "unspecified IPv4 -> 4-byte",
			host: net.ParseIP("1.1.1.1"),
			expected: net.ParseIP("1.1.1.1").To4(),
		},
		{
			name: "4-byte IPv4 -> 4-byte",
			host: net.ParseIP("1.1.1.2").To4(),
			expected: net.ParseIP("1.1.1.2").To4(),
		},
		{
			name: "16-byte IPv4 -> 4-byte",
			host: net.ParseIP("1.1.1.3").To16(),
			expected: net.ParseIP("1.1.1.3").To4(),
		},
		{
			name: "unspecified IPv6 -> 16-byte",
			host: net.ParseIP("3307::3f0f"),
			expected: net.ParseIP("3307::3f0f").To16(),
		},
		{
			name: "4-byte IPv6 -> 16-byte",
			host: net.ParseIP("3307::3f10"),
			expected: net.ParseIP("3307::3f10").To16(),
		},
		{
			name: "16-byte IPv6 -> 16-byte",
			host: net.ParseIP("3307::3f11"),
			expected: net.ParseIP("3307::3f11").To16(),
		},
	}

	for _, test := range tests {
		canonical := canonicalizeHosts([]net.IP{ test.host })
		assert.Equal(t, []net.IP{ test.expected }, canonical)
	}
}

func TestSortAndDedupHosts (t *testing.T) {
	tests := []struct{
		name string
		hosts []net.IP
		expected []net.IP
	}{
		{
			name: "de-dup and sort",
			hosts: []net.IP{
				net.ParseIP("1.1.1.1").To4(),
				net.ParseIP("3307::3f0f").To16(),
				net.ParseIP("1.1.1.2").To4(),
				net.ParseIP("1.1.1.0").To4(),
				net.ParseIP("3307::3f10").To16(),
				net.ParseIP("1.1.1.1").To4(),
				net.ParseIP("3307::3f0f").To16(),
			},
			expected: []net.IP{
				net.ParseIP("1.1.1.0").To4(),
				net.ParseIP("1.1.1.1").To4(),
				net.ParseIP("1.1.1.2").To4(),
				net.ParseIP("3307::3f0f").To16(),
				net.ParseIP("3307::3f10").To16(),
			},
		},
	}

	for _, test := range tests {
		prepared := sortAndDedupHosts(test.hosts)
		assert.Equal(t, test.expected, prepared)
	}
}

func TestCanonicalizeNetworks (t *testing.T) {
	tests := []struct{
		name string
		network net.IPNet
		expected net.IPNet
	}{
		{
			name: "unspecified IPv4 -> 4-byte",
			network: net.IPNet{
				IP: net.ParseIP("1.1.1.0"),
				Mask: net.CIDRMask(31,32),
			},
			expected: net.IPNet{
				IP: net.ParseIP("1.1.1.0").To4(),
				Mask: net.CIDRMask(31,32),
			},
		},
		{
			name: "network address corrected",
			network: net.IPNet{
				IP: net.ParseIP("1.1.1.1"),
				Mask: net.CIDRMask(31,32),
			},
			expected: net.IPNet{
				IP: net.ParseIP("1.1.1.0").To4(),
				Mask: net.CIDRMask(31,32),
			},
		},
	}

	for _, test := range tests {
		canonical := canonicalizeNetworks([]net.IPNet{ test.network })
		assert.Equal(t, []net.IPNet{ test.expected }, canonical)
	}
}

// func TestSortAndDedupNetworks (t *testing.T) {
// 	tests := []struct{
// 		name string
// 		networks []net.IP
// 		expected []net.IP
// 	}{
// 		{
// 			name: "de-dup and sort",
// 			networks: []net.IP{
// 				net.ParseIP("1.1.1.1").To4(),
// 				net.ParseIP("3307::3f0f").To16(),
// 				net.ParseIP("1.1.1.2").To4(),
// 				net.ParseIP("1.1.1.0").To4(),
// 				net.ParseIP("3307::3f10").To16(),
// 				net.ParseIP("1.1.1.1").To4(),
// 				net.ParseIP("3307::3f0f").To16(),
// 			},
// 			expected: []net.IP{
// 				net.ParseIP("1.1.1.0").To4(),
// 				net.ParseIP("1.1.1.1").To4(),
// 				net.ParseIP("1.1.1.2").To4(),
// 				net.ParseIP("3307::3f0f").To16(),
// 				net.ParseIP("3307::3f10").To16(),
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		prepared := sortAndDedupNetworks([]net.IP{ test.networks })
// 		assert.Equal(t, test.expected, prepared)
// 	}
// }

func TestTrySumHosts(t *testing.T) {
	tests := []struct{
		name string
		ips [2]string
		expected *net.IPNet
	}{
		{
			name: "different lengths",
			ips: [2]string{
				"1.1.1.1",
				"f2d4:126a:9eca:d7fc:6e1b:6846:5403:f201",
			},
			expected: nil,
		},
		{
			name: "same IPs",
			ips: [2]string{
				"1.1.1.0",
				"1.1.1.0",
			},
			expected: nil,
		},
		{
			name: "not in same /31",
			ips: [2]string{
				"1.1.1.1",
				"1.1.1.2",
			},
			expected: nil,
		},
		{
			name: "simple summarization",
			ips: [2]string{
				"1.1.1.0",
				"1.1.1.1",
			},
			expected: &net.IPNet{
				IP: canonicalIP(net.ParseIP("1.1.1.0")),
				Mask: net.CIDRMask(31, 32),
			},
		},
	}

	for _, test := range tests {
		a := canonicalIP(net.ParseIP(test.ips[0]))
		b := canonicalIP(net.ParseIP(test.ips[1]))
		sumNet := trySumHosts(a, b)
		assert.Equal(t, test.expected, sumNet, test.name)
	}
}

func TestSumRoutes (t *testing.T) {
	tests := []struct{
		name string
		ips []string
		expectedHosts []string
		expectedNetworks []string
	}{
		{
			name: "consecutive IPs, 2 full /31s",
			ips: []string{
				"1.1.1.2",
				"1.1.1.3",
				"1.1.1.4",
				"1.1.1.5",
			},
			expectedHosts: []string{},
			expectedNetworks: []string{
				"1.1.1.2/31",
				"1.1.1.4/31",
			},
		},
		{
			name: "consecutive IPs, 1 full /31",
			ips: []string{
				"1.1.1.1",
				"1.1.1.2",
				"1.1.1.3",
				"1.1.1.4",
			},
			expectedHosts: []string{
				"1.1.1.1",
				"1.1.1.4",
			},
			expectedNetworks: []string{
				"1.1.1.2/31",
			},
		},
		{
			name: "consecutive IPs in different /31s",
			ips: []string{
				"1.1.1.1",
				"1.1.1.2",
			},
			expectedHosts: []string{
				"1.1.1.1",
				"1.1.1.2",
			},
			expectedNetworks: []string{},
		},
		{
			name: "consecutive IPs in different /24s",
			ips: []string{
				"1.1.1.255",
				"1.1.2.0",
			},
			expectedHosts: []string{
				"1.1.1.255",
				"1.1.2.0",
			},
			expectedNetworks: []string{},
		},
		{
			name: "consecutive IPs create consecutive /31s",
			ips: []string{
				"1.1.1.0",
				"1.1.1.1",
				"1.1.1.2",
				"1.1.1.3",
			},
			expectedHosts: []string{},
			expectedNetworks: []string{"1.1.1.0/30"},
		},
		{
			name: "networks combine",
			ips: []string{
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
			expectedHosts: []string{
				"2.2.2.18",
			},
			expectedNetworks: []string{
				"2.2.2.16/31",
				"2.2.2.8/29",
			},
		},
	}

	for _, test := range tests {
		ips := make([]net.IP, len(test.ips))
		for i, s := range test.ips {
			ips[i] = canonicalIP(net.ParseIP(s))
		}

		networks, hosts := SumRoutes(ips, []net.IPNet{})

		assert.Equal(t, test.expectedHosts, stringifyHosts(hosts), "hosts: " + test.name)
		assert.Equal(t,	test.expectedNetworks, stringifyNetworks(networks), "networks: " + test.name)
	}
}

func stringifyNetworks(networks []net.IPNet) []string {
	stringified := make([]string, len(networks))
	for i, network := range networks {
		stringified[i] = network.String()
	}
	return stringified
}

func stringifyHosts(hosts []net.IP) []string {
	stringified := make([]string, len(hosts))
	for i, host := range hosts {
		stringified[i] = host.String()
	}
	return stringified
}
