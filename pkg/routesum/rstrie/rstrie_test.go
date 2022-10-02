package rstrie

import (
	"testing"

	"github.com/PatrickCronin/routesum/pkg/routesum/routetype"
	"github.com/stretchr/testify/assert"
)

func TestRSTrieV4InsertRoute(t *testing.T) { //nolint: funlen
	tests := []struct {
		name     string
		routes   []string
		expected *RSTrie[*routetype.V4]
	}{
		{
			name:   "add one child",
			routes: []string{"0.0.0.0/17"},
			expected: &RSTrie[*routetype.V4]{
				root: &node[*routetype.V4]{
					route:    routetype.MustParseV4String("0.0.0.0/17"),
					children: nil,
				},
			},
		},
		{
			name:   "add two children, completing the root node's subtrie",
			routes: []string{"0.0.0.0/1", "128.0.0.0/1"},
			expected: &RSTrie[*routetype.V4]{
				root: &node[*routetype.V4]{
					route:    routetype.MustParseV4String("0.0.0.0/0"),
					children: nil,
				},
			},
		},
		{
			name:   "split root, root is empty",
			routes: []string{"0.0.0.0/2", "192.0.0.0/2"},
			expected: &RSTrie[*routetype.V4]{
				root: &node[*routetype.V4]{
					route: routetype.MustParseV4String("0.0.0.0/0"),
					children: &[2]*node[*routetype.V4]{
						{
							route:    routetype.MustParseV4String("0.0.0.0/2"),
							children: nil,
						},
						{
							route:    routetype.MustParseV4String("192.0.0.0/2"),
							children: nil,
						},
					},
				},
			},
		},
		{
			name:   "split root, root is not empty",
			routes: []string{"0.0.0.0/2", "64.0.0.0/3"},
			expected: &RSTrie[*routetype.V4]{
				root: &node[*routetype.V4]{
					route: routetype.MustParseV4String("0.0.0.0/1"),
					children: &[2]*node[*routetype.V4]{
						{
							route:    routetype.MustParseV4String("0.0.0.0/2"),
							children: nil,
						},
						{
							route:    routetype.MustParseV4String("64.0.0.0/3"),
							children: nil,
						},
					},
				},
			},
		},
		{
			name:   "split root, traverse, and split internal",
			routes: []string{"0.0.0.0/1", "128.0.0.0/3", "192.0.0.0/3"}, // 0, 1-0-0, 1-1-0
			expected: &RSTrie[*routetype.V4]{
				root: &node[*routetype.V4]{
					route: routetype.MustParseV4String("0.0.0.0/0"),
					children: &[2]*node[*routetype.V4]{
						{
							route:    routetype.MustParseV4String("0.0.0.0/1"),
							children: nil,
						},
						{
							route: routetype.MustParseV4String("128.0.0.0/1"),
							children: &[2]*node[*routetype.V4]{
								{
									route:    routetype.MustParseV4String("128.0.0.0/3"),
									children: nil,
								},
								{
									route:    routetype.MustParseV4String("192.0.0.0/3"),
									children: nil,
								},
							},
						},
					},
				},
			},
		},
		{
			name:   "failing parent test",
			routes: []string{"192.0.2.1", "192.0.2.2", "192.0.2.3", "192.0.2.4"},
			expected: &RSTrie[*routetype.V4]{
				root: &node[*routetype.V4]{
					route: routetype.MustParseV4String("192.0.2.0/29"),
					children: &[2]*node[*routetype.V4]{
						{
							route: routetype.MustParseV4String("192.0.2.0/30"),
							children: &[2]*node[*routetype.V4]{
								{
									route:    routetype.MustParseV4String("192.0.2.1"),
									children: nil,
								},
								{
									route:    routetype.MustParseV4String("192.0.2.2/31"),
									children: nil,
								},
							},
						},
						{
							route:    routetype.MustParseV4String("192.0.2.4"),
							children: nil,
						},
					},
				},
			},
		},
		{
			name:   "covered routes are ignored",
			routes: []string{"0.0.0.0/1", "0.0.0.0/2"},
			expected: &RSTrie[*routetype.V4]{
				root: &node[*routetype.V4]{
					route:    routetype.MustParseV4String("0.0.0.0/1"),
					children: nil,
				},
			},
		},
		{
			name:   "route covering node replaces it",
			routes: []string{"0.0.0.0/2", "0.0.0.0/1"},
			expected: &RSTrie[*routetype.V4]{
				root: &node[*routetype.V4]{
					route:    routetype.MustParseV4String("0.0.0.0/1"),
					children: nil,
				},
			},
		},
		{
			name:   "completed subtries are simplified",
			routes: []string{"128.0.0.0/1", "64.0.0.0/2", "32.0.0.0/3", "0.0.0.0/3"},
			expected: &RSTrie[*routetype.V4]{
				root: &node[*routetype.V4]{
					route:    routetype.MustParseV4String("0.0.0.0/0"),
					children: nil,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trie := New[*routetype.V4]()

			for _, r := range test.routes {
				trie.InsertRoute(routetype.MustParseV4String(r))
			}

			assert.Equal(t, test.expected, trie, "got expected rstrie")
		})
	}
}

func TestRSTrieV4Contents(t *testing.T) { //nolint: funlen
	tests := []struct {
		name     string
		trie     RSTrie[*routetype.V4]
		expected []*routetype.V4
	}{
		{
			name: "complete trie",
			trie: RSTrie[*routetype.V4]{
				root: &node[*routetype.V4]{
					route:    routetype.MustParseV4String("0.0.0.0/0"),
					children: nil,
				},
			},
			expected: []*routetype.V4{routetype.MustParseV4String("0.0.0.0/0")},
		},
		{
			name: "empty trie",
			trie: RSTrie[*routetype.V4]{
				root: nil,
			},
			expected: []*routetype.V4(nil),
		},
		{
			name: "single zero-child trie",
			trie: RSTrie[*routetype.V4]{
				root: &node[*routetype.V4]{
					route:    routetype.MustParseV4String("0.0.0.0/1"),
					children: nil,
				},
			},
			expected: []*routetype.V4{routetype.MustParseV4String("0.0.0.0/1")},
		},
		{
			name: "single one-child trie",
			trie: RSTrie[*routetype.V4]{
				root: &node[*routetype.V4]{
					route:    routetype.MustParseV4String("128.0.0.0/1"),
					children: nil,
				},
			},
			expected: []*routetype.V4{routetype.MustParseV4String("128.0.0.0/1")},
		},
		{
			name: "two-level trie",
			trie: RSTrie[*routetype.V4]{
				root: &node[*routetype.V4]{
					route: routetype.MustParseV4String("0.0.0.0/2"),
					children: &[2]*node[*routetype.V4]{
						{
							route:    routetype.MustParseV4String("0.0.0.0/3"),
							children: nil,
						},
						{
							route: routetype.MustParseV4String("32.0.0.0/4"),
						},
					},
				},
			},
			expected: []*routetype.V4{
				routetype.MustParseV4String("0.0.0.0/3"),
				routetype.MustParseV4String("32.0.0.0/4"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.trie.Contents(), "got expected contents")
		})
	}
}

func TestRSTrieV4MemUsage(t *testing.T) {
	tests := []struct {
		name                     string
		entries                  []string
		expectedNumInternalNodes uint
		expectedNumLeafNodes     uint
	}{
		{
			name:                     "new trie",
			expectedNumInternalNodes: 0,
			expectedNumLeafNodes:     0,
		},
		{
			name: "one item",
			entries: []string{
				"192.0.2.1",
			},
			expectedNumInternalNodes: 0,
			expectedNumLeafNodes:     1,
		},
		{
			name: "two items, summarized",
			entries: []string{
				"192.0.2.1",
				"192.0.2.0",
			},
			expectedNumInternalNodes: 0,
			expectedNumLeafNodes:     1,
		},
		{
			name: "two items, unsummarized",
			entries: []string{
				"192.0.2.1",
				"192.0.2.2",
			},
			expectedNumInternalNodes: 1,
			expectedNumLeafNodes:     2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trie := New[*routetype.V4]()

			for _, entry := range test.entries {
				ipv4 := routetype.MustParseV4String(entry)
				trie.InsertRoute(ipv4)
			}

			numInternalNodes, numLeafNodes, _, _ := trie.MemUsage()
			assert.Equal(t, test.expectedNumInternalNodes, numInternalNodes, "num internal nodes")
			assert.Equal(t, test.expectedNumLeafNodes, numLeafNodes, "num leaf nodes")
		})
	}
}
