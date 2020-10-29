package routesum

import (
	"bytes"
	"sort"
)

type sortableNetworkSlice []safeRepNet

func newSortableNetworkSlice(srNets []safeRepNet) sortableNetworkSlice {
	return sortableNetworkSlice(srNets)
}

func (sns sortableNetworkSlice) Len() int {
	return len(sns)
}

func (sns sortableNetworkSlice) Less(i, j int) bool {
	// IPv4 before IPv6
	if len(sns[i].IP) != len(sns[j].IP) {
		return len(sns[i].IP) < len(sns[j].IP)
	}

	// networks with fewer hosts before networks with more hosts
	maskCmp := bytes.Compare(sns[j].Mask, sns[i].Mask)
	if maskCmp != 0 {
		return maskCmp < 0
	}

	// networks at "lower" addresses before networks at "higher" addresses
	return bytes.Compare(sns[i].IP.To16(), sns[j].IP.To16()) < 0
}

func (sns sortableNetworkSlice) Swap(i, j int) {
	sns[i], sns[j] = sns[j], sns[i]
}

func sortNetworksFromBigToSmall(srNets []safeRepNet) []safeRepNet {
	sns := newSortableNetworkSlice(srNets)
	sort.Sort(sort.Reverse(sns))
	return []safeRepNet(sns)
}

func sortNetworksFromSmallToBig(srNets []safeRepNet) []safeRepNet {
	sns := newSortableNetworkSlice(srNets)
	sort.Sort(sns)
	return []safeRepNet(sns)
}
