// Package rstrie provides a datatype that supports building a space-efficient summary of networks and IPs.
package rstrie

import (
	"bytes"
	"container/list"

	"github.com/PatrickCronin/routesum/pkg/routesum/bitslice"
)

// RSTrie is a radix-like trie of radix 2 whose stored "words" are the binary representations of networks and IPs. An
// optimization rstrie makes over a generic radix tree is that since routes covered by other routes don't need to be
// stored, each node in the trie will have either 0 or 2 children; never 1.
type RSTrie struct {
	root *node
}

type node struct {
	children *[2]*node
	bits     bitslice.BitSlice
}

// NewRSTrie returns an initialized RSTrie for use
func NewRSTrie() *RSTrie {
	return &RSTrie{
		root: nil,
	}
}

// InsertRoute inserts a new BitSlice into the trie. Each insert results in a space-optimized trie structure
// representing its contents. If a route being inserted is already covered by an existing route, it's simply ignored. If
// a route being inserted covers one or more routes already in the trie, those nodes are removed and replaced by the new
// route.
func (t *RSTrie) InsertRoute(routeBits bitslice.BitSlice) {
	// If the trie has no root node, simply create one to store the new route
	if t.root == nil {
		t.root = &node{
			bits:     routeBits,
			children: nil,
		}
		return
	}

	t.root.insertRoute(&t.root, routeBits)
}

func (n *node) isLeaf() bool {
	return n.children == nil
}

// parent is a **node so that we can change what the parent is pointing to if we need to!
func (n *node) insertRoute(parent **node, remainingRouteBits bitslice.BitSlice) bool {
	remainingRouteBitsLen := len(remainingRouteBits)
	curNodeBitsLen := len(n.bits)

	// Does the requested route cover the current node? If so, update the current node.
	if remainingRouteBitsLen <= curNodeBitsLen && bytes.HasPrefix(n.bits, remainingRouteBits) {
		n.bits = remainingRouteBits
		n.children = nil
		return true
	}

	if curNodeBitsLen <= remainingRouteBitsLen && bytes.HasPrefix(remainingRouteBits, n.bits) {
		// Does the current node cover the requested route? If so, we're done.
		if n.isLeaf() {
			return false
		}

		// Otherwise, we traverse to the correct child.
		whichChild := remainingRouteBits[curNodeBitsLen]
		if n.children[whichChild].insertRoute(&n.children[whichChild], remainingRouteBits[curNodeBitsLen:]) {
			return n.maybeRemoveRedundantChildren()
		}

		return false
	}

	// Otherwise the requested route diverges from the current node. We'll need to split the current node.

	// As an optimization, if the split would result in a new node whose children represent a complete subtrie, we
	// just update the current node, instead of allocating new nodes and optimizing them away immediately after.
	if n.isLeaf() &&
		curNodeBitsLen == remainingRouteBitsLen &&
		commonPrefixLen(n.bits, remainingRouteBits) == len(n.bits)-1 {
		n.bits = n.bits[:len(n.bits)-1]
		n.children = nil
		return true
	}

	*parent = splitNodeForRoute(n, remainingRouteBits)
	return n.maybeRemoveRedundantChildren()
}

func commonPrefixLen(a, b bitslice.BitSlice) int {
	i := 0
	maxLen := min(len(a), len(b))
	for ; i < maxLen; i++ {
		if a[i] != b[i] {
			break
		}
	}

	return i
}

func splitNodeForRoute(oldNode *node, routeBits bitslice.BitSlice) *node {
	commonBitsLen := commonPrefixLen(oldNode.bits, routeBits)
	commonBits := oldNode.bits[:commonBitsLen]

	routeNode := &node{
		bits:     routeBits[commonBitsLen:],
		children: nil,
	}
	oldNode.bits = oldNode.bits[commonBitsLen:]

	newNode := &node{
		bits:     commonBits,
		children: &[2]*node{},
	}
	newNode.children[routeNode.bits[0]] = routeNode
	newNode.children[oldNode.bits[0]] = oldNode

	return newNode
}

// A node's children are redundant if they, taken together, represent a complete subtrie from the
// node's perspective. This situation can be represented more simply as the node having a nil
// children pointer.
func (n *node) maybeRemoveRedundantChildren() bool {
	if n.isLeaf() {
		return false
	}

	if !n.children[0].isLeaf() || !n.children[1].isLeaf() {
		return false
	}

	if len(n.children[0].bits) != 1 || len(n.children[1].bits) != 1 {
		return false
	}

	n.children = nil
	return true
}

type traversalStep struct {
	n                  *node
	precedingRouteBits bitslice.BitSlice
}

// Contents returns the BitSlices contained in the RSTrie.
func (t *RSTrie) Contents() []bitslice.BitSlice {
	// If the trie is empty
	if t.root == nil {
		return []bitslice.BitSlice{}
	}

	// Otherwise
	remainingSteps := list.New()
	remainingSteps.PushFront(traversalStep{
		n:                  t.root,
		precedingRouteBits: bitslice.BitSlice{},
	})

	contents := []bitslice.BitSlice{}
	for remainingSteps.Len() > 0 {
		step := remainingSteps.Remove(remainingSteps.Front()).(traversalStep)

		stepRouteBits := bitslice.BitSlice{}
		stepRouteBits = append(stepRouteBits, step.precedingRouteBits...)
		stepRouteBits = append(stepRouteBits, step.n.bits...)

		if step.n.isLeaf() {
			contents = append(contents, stepRouteBits)
		} else {
			remainingSteps.PushFront(traversalStep{
				n:                  step.n.children[1],
				precedingRouteBits: stepRouteBits,
			})
			remainingSteps.PushFront(traversalStep{
				n:                  step.n.children[0],
				precedingRouteBits: stepRouteBits,
			})
		}
	}

	return contents
}
