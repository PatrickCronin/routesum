// Package routesum summarizes a list of IPs and networks to its shortest form.
package routesum

import (
	"strings"

	"github.com/PatrickCronin/routesum/pkg/routesum/routetype"
	"github.com/PatrickCronin/routesum/pkg/routesum/rstrie"
)

// RouteSum has methods supporting route summarization of networks and hosts
type RouteSum struct {
	v4 *rstrie.RSTrie[*routetype.V4]
	v6 *rstrie.RSTrie[*routetype.V6]
}

// New returns an initialized RouteSum object
func New() *RouteSum {
	rs := new(RouteSum)
	rs.v4 = rstrie.New[*routetype.V4]()
	rs.v6 = rstrie.New[*routetype.V6]()

	return rs
}

// InsertFromString adds either a string-formatted network or IP to the summary
func (rs *RouteSum) InsertFromString(s string) error {
	if strings.Contains(s, ":") {
		r, err := routetype.ParseV6String(s)
		if err != nil {
			return err //nolint: wrapcheck
		}

		rs.v6.InsertRoute(r)
	} else {
		r, err := routetype.ParseV4String(s)
		if err != nil {
			return err //nolint: wrapcheck
		}

		rs.v4.InsertRoute(r)
	}

	return nil
}

// SummaryStrings returns a summary of all received routes as a string slice.
func (rs *RouteSum) SummaryStrings() ([]string, error) {
	strs := []string{}

	for _, r := range rs.v4.Contents() {
		str, err := r.String()
		if err != nil {
			return nil, err //nolint: wrapcheck
		}

		strs = append(strs, str)
	}

	for _, r := range rs.v6.Contents() {
		str, err := r.String()
		if err != nil {
			return nil, err //nolint: wrapcheck
		}

		strs = append(strs, str)
	}

	return strs, nil
}

// MemUsage provides information about memory usage.
func (rs *RouteSum) MemUsage() (uint, uint, uintptr, uintptr) {
	v4NumInternalNodes, v4NumLeafNodes, v4InternalNodesTotalSize, v4LeafNodesTotalSize := rs.v4.MemUsage()
	v6NumInternalNodes, v6NumLeafNodes, v6InternalNodesTotalSize, v6LeafNodesTotalSize := rs.v6.MemUsage()
	return v4NumInternalNodes + v6NumInternalNodes,
		v4NumLeafNodes + v6NumLeafNodes,
		v4InternalNodesTotalSize + v6InternalNodesTotalSize,
		v4LeafNodesTotalSize + v6LeafNodesTotalSize
}
