package routesum

import (
	"bytes"
	"fmt"
	"net"
	"strings"
)

// Strings summarizes routes from a list of (string representations of) networks
// and IPs. Networks should be specified using CIDR notation.
func Strings(strs []string) ([]string, error) {
	// Parse and validate
	var srNets []safeRepNet
	var srIPs []safeRepIP
	for _, s := range strs {
		if strings.Index(s, "/") != -1 {
			srNet, err := newSafeRepNetFromString(s)
			if err != nil {
				return nil, fmt.Errorf("validate network: %w", err)
			}

			srNets = append(srNets, *srNet)
		} else {
			srIP, err := newSafeRepIPFromString(s)
			if err != nil {
				return nil, fmt.Errorf("validate IP: %w", err)
			}
			srIPs = append(srIPs, srIP)
		}
	}

	// Summarize
	summarizedNetworks, remainingIPs, err := networksAndIPs(srNets, srIPs)
	if err != nil {
		return nil, fmt.Errorf("summarize networks and IPs: %w", err)
	}

	// Provide results in the same format we got them
	summarizedStrs := make([]string, len(summarizedNetworks)+len(remainingIPs))
	for i, srIP := range remainingIPs {
		summarizedStrs[i] = srIP.String()
	}
	numIPs := len(remainingIPs)
	for i, srNet := range summarizedNetworks {
		summarizedStrs[i+numIPs] = srNet.String()
	}
	return summarizedStrs, nil
}

// NetworksAndIPs summarizes routes from a set of []net.IPNet and []net.IP
// objects.
func NetworksAndIPs(
	networks []net.IPNet,
	ips []net.IP,
) ([]net.IPNet, []net.IP, error) {
	// Validate
	var srNets []safeRepNet
	for _, network := range networks {
		srNet, err := newSafeRepNetFromNetIPNet(network)
		if err != nil {
			return nil, nil, fmt.Errorf("validate network: %w", err)
		}
		srNets = append(srNets, *srNet)
	}

	var srIPs []safeRepIP
	for _, ip := range ips {
		srIP, err := newSafeRepIPFromNetIP(ip)
		if err != nil {
			return nil, nil, fmt.Errorf("validate IP: %w", err)
		}
		srIPs = append(srIPs, srIP)
	}

	// Summarize
	summarizedNetworks, remainingIPs, err := networksAndIPs(srNets, srIPs)
	if err != nil {
		return nil, nil, fmt.Errorf("summarize networks and IPs: %w", err)
	}

	// Provide results in the same format we got them
	sumNets := make([]net.IPNet, len(summarizedNetworks))
	for i, sumNet := range summarizedNetworks {
		sumNets[i] = net.IPNet(sumNet)
	}

	remIPs := make([]net.IP, len(remainingIPs))
	for i, remIP := range remainingIPs {
		remIPs[i] = net.IP(remIP)
	}

	return sumNets, remIPs, nil
}

func networksAndIPs(
	srNets []safeRepNet,
	srIPs []safeRepIP,
) ([]safeRepNet, []safeRepIP, error) {
	// To simplify implementation, we translate any IPs to networks with a
	// subnet mask indicating 0 hosts.
	zeroHostMask := map[int]net.IPMask{
		net.IPv4len: net.CIDRMask(32, 32),
		net.IPv6len: net.CIDRMask(128, 128),
	}

	zeroHostNets := make([]safeRepNet, len(srIPs))
	for i, srIP := range srIPs {
		zeroHostNets[i] = safeRepNet{
			IP:   net.IP(srIP),
			Mask: zeroHostMask[len(srIP)],
		}
	}

	allNets := append(zeroHostNets, srNets...)
	allCleanedNets := removeContainedNetworks(allNets)

	summarizedNetworks, err := summarizeNetworks(allCleanedNets)
	if err != nil {
		return nil, nil, fmt.Errorf("summarize networks: %w", err)
	}

	// Re-interpret the zero-host networks as IPs
	var sumNets []safeRepNet
	var sumIPs []safeRepIP
	for _, srNet := range summarizedNetworks {
		if bytes.Compare(zeroHostMask[len(srNet.IP)], srNet.Mask) == 0 {
			sumIPs = append(sumIPs, safeRepIP(srNet.IP))
		} else {
			sumNets = append(sumNets, srNet)
		}
	}

	return sumNets, sumIPs, nil
}

// We remove any networks that are fully contained by another in the list. E.g.
// if 192.0.2.0/24 and 192.2.0.0/23 are both in the list, remove the former
// as it's fully contained by the latter.
func removeContainedNetworks(networks []safeRepNet) []safeRepNet {
	candidateNets := sortNetworksFromBigToSmall(networks)
	var nonContainedNets []safeRepNet
candidate:
	for _, candidate := range candidateNets {
		for _, nonContainedNet := range nonContainedNets {
			if bytes.Compare(
				candidate.IP.Mask(nonContainedNet.Mask),
				nonContainedNet.IP,
			) == 0 {
				continue candidate
			}
		}

		nonContainedNets = append(nonContainedNets, candidate)
	}

	return nonContainedNets
}

func summarizeNetworks(srNets []safeRepNet) ([]safeRepNet, error) {
	thisRound := srNets
	var lastRound []safeRepNet
	for len(thisRound) != len(lastRound) { // Something was summarized
		lastRound = thisRound
		thisRound = summarizeNetworksOneRound(lastRound)
	}

	return thisRound, nil
}

func summarizeNetworksOneRound(srNets []safeRepNet) []safeRepNet {
	sortedSRNets := sortNetworksFromSmallToBig(srNets)

	var summary []safeRepNet
	numNets := len(sortedSRNets)
	for i := 0; i < numNets; i++ {
		if i < numNets-1 {
			sum := trySumNetworks(sortedSRNets[i], sortedSRNets[i+1])
			if sum != nil {
				summary = append(summary, *sum)
				i++
				continue
			}
		}

		summary = append(summary, sortedSRNets[i])
	}

	return summary
}

func trySumNetworks(a, b safeRepNet) *safeRepNet {
	// IPs from different families cannot be summarized
	if len(a.IP) != len(b.IP) {
		return nil
	}

	// IPs with different masks cannot be summarized
	if bytes.Compare(a.Mask, b.Mask) != 0 {
		return nil
	}

	// If the networks' base IPs are the same, there's nothing to summarize
	// because we've already asserted that no networks are covered by others.
	if bytes.Compare(a.IP, b.IP) == 0 {
		return nil
	}

	ones, bits := a.Mask.Size()
	if ones == 0 {
		return nil
	}

	sumMask := net.CIDRMask(ones-1, bits)
	networkA := a.IP.Mask(sumMask)

	if bytes.Compare(networkA, b.IP.Mask(sumMask)) != 0 {
		return nil
	}

	return &safeRepNet{
		IP:   networkA,
		Mask: sumMask,
	}
}
