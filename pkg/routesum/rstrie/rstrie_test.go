package rstrie

import (
	"testing"

	"github.com/PatrickCronin/routesum/pkg/routesum/bitvector"
	"github.com/stretchr/testify/assert"
)

func TestRSTrieInsertRoute(t *testing.T) { //nolint: funlen
	tests := []struct {
		name     string
		routes   []bitvector.BitVector
		expected *RSTrie
	}{
		{
			name:   "add one child",
			routes: []bitvector.BitVector{bitvector.New([]byte{0x0}, 1)},
			expected: &RSTrie{
				root: &node{
					bits:     bitvector.New([]byte{0x0}, 1),
					children: nil,
				},
			},
		},
		{
			name:   "add two children, completing the root node's subtrie",
			routes: []bitvector.BitVector{bitvector.New([]byte{0x0}, 1), bitvector.New([]byte{0x80}, 1)},
			expected: &RSTrie{root: &node{
				bits:     bitvector.BitVector{},
				children: nil,
			}},
		},
		{
			name: "split root, root is empty",
			routes: []bitvector.BitVector{
				bitvector.New([]byte{0x0}, 2),
				bitvector.New([]byte{0xc0}, 2),
			},
			expected: &RSTrie{
				root: &node{
					bits: bitvector.BitVector{},
					children: &[2]*node{
						0: {bits: bitvector.New([]byte{0x0}, 2)},
						1: {bits: bitvector.New([]byte{0xc0}, 2)},
					},
				},
			},
		},
		{
			name: "split root, root is not empty",
			routes: []bitvector.BitVector{
				bitvector.New([]byte{0x0}, 2),
				bitvector.New([]byte{0x40}, 3),
			},
			expected: &RSTrie{
				root: &node{
					bits: bitvector.New([]byte{0x0}, 1),
					children: &[2]*node{
						0: {bits: bitvector.New([]byte{0x0}, 1)},
						1: {bits: bitvector.New([]byte{0x80}, 2)},
					},
				},
			},
		},
		{
			name: "split root, traverse, and split internal",
			routes: []bitvector.BitVector{
				bitvector.New([]byte{0x0}, 1),
				bitvector.New([]byte{0x80}, 3),
				bitvector.New([]byte{0xc0}, 3),
			},
			expected: &RSTrie{
				root: &node{
					bits: bitvector.BitVector{},
					children: &[2]*node{
						0: {bits: bitvector.New([]byte{0x0}, 1)},
						1: {
							bits: bitvector.New([]byte{0x80}, 1),
							children: &[2]*node{
								0: {bits: bitvector.New([]byte{0x0}, 2)},
								1: {bits: bitvector.New([]byte{0x80}, 2)},
							},
						},
					},
				},
			},
		},
		{
			name: "covered routes are ignored",
			routes: []bitvector.BitVector{
				bitvector.New([]byte{0x0}, 1),
				bitvector.New([]byte{0x0}, 2),
			},
			expected: &RSTrie{
				root: &node{
					bits:     bitvector.New([]byte{0x0}, 1),
					children: nil,
				},
			},
		},
		{
			name: "route covering node replaces it",
			routes: []bitvector.BitVector{
				bitvector.New([]byte{0x0}, 2),
				bitvector.New([]byte{0x0}, 1),
			},
			expected: &RSTrie{
				root: &node{
					bits:     bitvector.New([]byte{0x0}, 1),
					children: nil,
				},
			},
		},
		{
			name: "completed subtries are simpliflied",
			routes: []bitvector.BitVector{
				bitvector.New([]byte{0x80}, 1),
				bitvector.New([]byte{0x40}, 2),
				bitvector.New([]byte{0x20}, 3),
				bitvector.New([]byte{0x0}, 3),
			},
			expected: &RSTrie{root: &node{
				bits:     bitvector.BitVector{},
				children: nil,
			}},
		},
		{
			name: "completed subtries are simplified when new route covers current",
			routes: []bitvector.BitVector{
				bitvector.New([]byte{0x0}, 2),
				bitvector.New([]byte{0x60}, 3),
				bitvector.New([]byte{0x40}, 2),
			},
			expected: &RSTrie{root: &node{
				bits:     bitvector.New([]byte{0x0}, 1),
				children: nil,
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trie := NewRSTrie()

			for _, route := range test.routes {
				trie.InsertRoute(route)
			}

			assert.Equal(t, test.expected, trie, "got expected rstrie")
		})
	}
}

func TestRSTrieContents(t *testing.T) { //nolint: funlen
	tests := []struct {
		name     string
		trie     RSTrie
		expected []bitvector.BitVector
	}{
		{
			name: "complete trie",
			trie: RSTrie{
				root: &node{
					bits:     bitvector.BitVector{},
					children: nil,
				},
			},
			expected: []bitvector.BitVector{{}},
		},
		{
			name: "empty trie",
			trie: RSTrie{
				root: nil,
			},
			expected: []bitvector.BitVector{},
		},
		{
			name: "single zero-child trie",
			trie: RSTrie{
				root: &node{
					bits:     bitvector.New([]byte{0x0}, 1),
					children: nil,
				},
			},
			expected: []bitvector.BitVector{bitvector.New([]byte{0x0}, 1)},
		},
		{
			name: "single one-child trie",
			trie: RSTrie{
				root: &node{
					bits:     bitvector.New([]byte{0x80}, 1),
					children: nil,
				},
			},
			expected: []bitvector.BitVector{bitvector.New([]byte{0x80}, 1)},
		},
		{
			name: "two-level trie",
			trie: RSTrie{
				root: &node{
					bits: bitvector.New([]byte{0x0}, 2),
					children: &[2]*node{
						0: {bits: bitvector.New([]byte{0x0}, 1)},
						1: {bits: bitvector.New([]byte{0x80}, 2)},
					},
				},
			},
			expected: []bitvector.BitVector{
				bitvector.New([]byte{0x0}, 3),
				bitvector.New([]byte{0x20}, 4),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.trie.Contents(), "got expected bits")
		})
	}
}
