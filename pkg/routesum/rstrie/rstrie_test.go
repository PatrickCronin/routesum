package rstrie

import (
	"testing"

	"github.com/PatrickCronin/routesum/pkg/routesum/bitslice"
	"github.com/stretchr/testify/assert"
)

func TestCommonPrefixLen(t *testing.T) {
	tests := []struct {
		name     string
		a, b     bitslice.BitSlice
		expected int
	}{
		{
			name:     "differing first bit",
			a:        bitslice.BitSlice{0},
			b:        bitslice.BitSlice{1},
			expected: 0,
		},
		{
			name:     "differing second bit",
			a:        bitslice.BitSlice{0, 0},
			b:        bitslice.BitSlice{0, 1},
			expected: 1,
		},
		{
			name:     "nothing different",
			a:        bitslice.BitSlice{0, 0, 0, 1},
			b:        bitslice.BitSlice{0, 0, 0, 1},
			expected: 4,
		},
	}

	for _, test := range tests {
		assert.Equal(
			t,
			test.expected,
			commonPrefixLen(test.a, test.b),
			test.name,
		)
	}
}

func TestRSTrieInsertRoute(t *testing.T) { //nolint: funlen
	tests := []struct {
		name     string
		routes   []bitslice.BitSlice
		expected *RSTrie
	}{
		{
			name:   "add one child",
			routes: []bitslice.BitSlice{{0}},
			expected: &RSTrie{
				root: &node{
					bits:     bitslice.BitSlice{0},
					children: nil,
				},
			},
		},
		{
			name:   "add two children, completing the root node's subtrie",
			routes: []bitslice.BitSlice{{0}, {1}},
			expected: &RSTrie{root: &node{
				bits:     bitslice.BitSlice{},
				children: nil,
			}},
		},
		{
			name:   "split root, root is empty",
			routes: []bitslice.BitSlice{{0, 0}, {1, 1}},
			expected: &RSTrie{
				root: &node{
					bits: bitslice.BitSlice{},
					children: &[2]*node{
						0: {bits: bitslice.BitSlice{0, 0}},
						1: {bits: bitslice.BitSlice{1, 1}},
					},
				},
			},
		},
		{
			name:   "split root, root is not empty",
			routes: []bitslice.BitSlice{{0, 0}, {0, 1, 0}},
			expected: &RSTrie{
				root: &node{
					bits: bitslice.BitSlice{0},
					children: &[2]*node{
						0: {bits: bitslice.BitSlice{0}},
						1: {bits: bitslice.BitSlice{1, 0}},
					},
				},
			},
		},
		{
			name:   "split root, traverse, and split internal",
			routes: []bitslice.BitSlice{{0}, {1, 0, 0}, {1, 1, 0}},
			expected: &RSTrie{
				root: &node{
					bits: bitslice.BitSlice{},
					children: &[2]*node{
						0: {bits: bitslice.BitSlice{0}},
						1: {
							bits: bitslice.BitSlice{1},
							children: &[2]*node{
								0: {bits: bitslice.BitSlice{0, 0}},
								1: {bits: bitslice.BitSlice{1, 0}},
							},
						},
					},
				},
			},
		},
		{
			name:   "covered routes are ignored",
			routes: []bitslice.BitSlice{{0}, {0, 0}},
			expected: &RSTrie{
				root: &node{
					bits:     bitslice.BitSlice{0},
					children: nil,
				},
			},
		},
		{
			name:   "route covering node replaces it",
			routes: []bitslice.BitSlice{{0, 0}, {0}},
			expected: &RSTrie{
				root: &node{
					bits:     bitslice.BitSlice{0},
					children: nil,
				},
			},
		},
		{
			name: "completed subtries are simpliflied",
			routes: []bitslice.BitSlice{
				{1},
				{0, 1},
				{0, 0, 1},
				{0, 0, 0},
			},
			expected: &RSTrie{root: &node{
				bits:     bitslice.BitSlice{},
				children: nil,
			}},
		},
		{
			name: "completed subtries are simplified when new route covers current",
			routes: []bitslice.BitSlice{
				{0, 0},
				{0, 1, 1},
				{0, 1},
			},
			expected: &RSTrie{root: &node{
				bits:     bitslice.BitSlice{0},
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
		expected []bitslice.BitSlice
	}{
		{
			name: "complete trie",
			trie: RSTrie{
				root: &node{
					bits:     nil,
					children: nil,
				},
			},
			expected: []bitslice.BitSlice{{}},
		},
		{
			name: "empty trie",
			trie: RSTrie{
				root: nil,
			},
			expected: []bitslice.BitSlice{},
		},
		{
			name: "single zero-child trie",
			trie: RSTrie{
				root: &node{
					bits:     bitslice.BitSlice{0},
					children: nil,
				},
			},
			expected: []bitslice.BitSlice{{0}},
		},
		{
			name: "single one-child trie",
			trie: RSTrie{
				root: &node{
					bits:     bitslice.BitSlice{1},
					children: nil,
				},
			},
			expected: []bitslice.BitSlice{{1}},
		},
		{
			name: "two-level trie",
			trie: RSTrie{
				root: &node{
					bits: bitslice.BitSlice{0, 0},
					children: &[2]*node{
						0: {bits: bitslice.BitSlice{0}},
						1: {bits: bitslice.BitSlice{1, 0}},
					},
				},
			},
			expected: []bitslice.BitSlice{{0, 0, 0}, {0, 0, 1, 0}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.trie.Contents(), "got expected bits")
		})
	}
}
