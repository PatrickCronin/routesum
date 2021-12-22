package rstree

import (
	"testing"

	"github.com/PatrickCronin/routesum/pkg/routesum/bitslice"
	"github.com/stretchr/testify/assert"
)

func TestRSTreeInsertRoute(t *testing.T) { //nolint: funlen
	tests := []struct {
		name     string
		routes   []bitslice.BitSlice
		expected *RSTree
	}{
		{
			name:   "add one child",
			routes: []bitslice.BitSlice{{0}},
			expected: &RSTree{
				root: &node{
					children: &[2]*node{
						0: new(node),
					},
				},
			},
		},
		{
			name:   "add two children, completing the root node's subtree",
			routes: []bitslice.BitSlice{{0}, {1}},
			expected: &RSTree{
				root: &node{children: nil},
			},
		},
		{
			name:   "covered routes are ignored",
			routes: []bitslice.BitSlice{{0}, {0, 0}},
			expected: &RSTree{
				root: &node{
					children: &[2]*node{
						0: new(node),
					},
				},
			},
		},
		{
			name:   "route covering node replaces it",
			routes: []bitslice.BitSlice{{0, 0}, {0}},
			expected: &RSTree{
				root: &node{
					children: &[2]*node{
						0: new(node),
					},
				},
			},
		},
		{
			name: "completed subtrees are simpliflied",
			routes: []bitslice.BitSlice{
				{1},
				{0, 1},
				{0, 0, 1},
				{0, 0, 0},
			},
			expected: &RSTree{
				root: &node{children: nil},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tree := NewRSTree()

			for _, route := range test.routes {
				tree.InsertRoute(route)
			}

			assert.Equal(t, test.expected, tree, "got expected rstree")
		})
	}
}

func TestRSTreeContents(t *testing.T) { //nolint: funlen
	tests := []struct {
		name     string
		tree     RSTree
		expected []bitslice.BitSlice
	}{
		{
			name: "complete tree",
			tree: RSTree{
				root: &node{children: nil},
			},
			expected: []bitslice.BitSlice{{}},
		},
		{
			name: "empty tree",
			tree: RSTree{
				root: nil,
			},
			expected: []bitslice.BitSlice{},
		},
		{
			name: "single one-child tree (0)",
			tree: RSTree{
				root: &node{
					children: &[2]*node{
						0: new(node),
					},
				},
			},
			expected: []bitslice.BitSlice{{0}},
		},
		{
			name: "single one-child tree (1)",
			tree: RSTree{
				root: &node{
					children: &[2]*node{
						1: new(node),
					},
				},
			},
			expected: []bitslice.BitSlice{{1}},
		},
		{
			name: "multi-level tree",
			tree: RSTree{
				root: &node{
					children: &[2]*node{
						0: {
							children: &[2]*node{
								0: {
									children: &[2]*node{
										0: new(node),
										1: {
											children: &[2]*node{
												0: new(node),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []bitslice.BitSlice{{0, 0, 0}, {0, 0, 1, 0}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.tree.Contents(), "got expected bits")
		})
	}
}
