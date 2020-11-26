package routesum

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSortableNetworkSlice(t *testing.T) {
	networkStrs := []string{
		"203.0.113.4/30",
		"198.51.100.0/24",
		"192.0.2.0/26",
		"192.0.2.64/26",
		"192.0.2.128/26",
		"192.0.2.192/26",
		"::ffff:192.0.2.0/128",
		"2600::/64",
	}

	srNets := make([]safeRepNet, len(networkStrs))
	for i, s := range networkStrs {
		srNet, err := newSafeRepNetFromString(s)
		require.NoError(t, err, "parse network")
		srNets[i] = *srNet
	}

	bigToSmall := sortNetworksFromBigToSmall(srNets)
	bigToSmallStrs := make([]string, len(bigToSmall))
	for i, srNet := range bigToSmall {
		bigToSmallStrs[i] = srNet.String()
	}
	assert.Equal(
		t,
		[]string{
			"2600::/64",
			"::ffff:192.0.2.0/128",
			"198.51.100.0/24",
			"192.0.2.192/26",
			"192.0.2.128/26",
			"192.0.2.64/26",
			"192.0.2.0/26",
			"203.0.113.4/30",
		},
		bigToSmallStrs,
		"big to small",
	)

	smallToBig := sortNetworksFromSmallToBig(srNets)
	smallToBigStrs := make([]string, len(smallToBig))
	for i, srNet := range smallToBig {
		smallToBigStrs[i] = srNet.String()
	}
	assert.Equal(
		t,
		[]string{
			"203.0.113.4/30",
			"192.0.2.0/26",
			"192.0.2.64/26",
			"192.0.2.128/26",
			"192.0.2.192/26",
			"198.51.100.0/24",
			"::ffff:192.0.2.0/128",
			"2600::/64",
		},
		smallToBigStrs,
		"small to big",
	)
}
