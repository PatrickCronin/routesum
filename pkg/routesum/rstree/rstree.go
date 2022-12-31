// Package rstree provides a datatype that supports building a space-efficient summary of networks
// and IPs.
package rstree

import (
	"container/list"
	"unsafe"

	"github.com/PatrickCronin/routesum/pkg/routesum/bitslice"
)

type node struct {
	children *[2]*node
}

func (n *node) isLeaf() bool {
	return n.children == nil
}

// RSTree is a binary tree that supports the storage and retrieval of networks and IPs for the
// purpose of route summarization.
type RSTree struct {
	root *node
}

// NewRSTree returns an initialized RSTree for use
func NewRSTree() *RSTree {
	return &RSTree{
		root: nil,
	}
}

// InsertRoute inserts a new BitSlice into the tree. Each insert results in a space-optimized tree
// structure. If a route being inserted is already covered by an existing route, it's simply
// ignored. If a route being inserted covers one or more routes already stored, those routes are
// replaced.
func (t *RSTree) InsertRoute(routeBits bitslice.BitSlice) {
	// If the tree has no root node, create one.
	if t.root == nil {
		t.root = &node{children: nil}

		if len(routeBits) > 0 {
			t.root.children = new([2]*node)
		}
	}

	t.root.insertRoute(routeBits)
}

// insertRoute returns whether or not a change was made to the tree that might require upwards
// optimization.
func (n *node) insertRoute(remainingRouteBits bitslice.BitSlice) bool {
	// Does the current node cover the requested route? If so, we're done.
	if n.isLeaf() {
		return false
	}

	// Does the requested route cover the current node? If so, update the current node.
	remainingRouteBitsLen := len(remainingRouteBits)
	if remainingRouteBitsLen == 0 {
		n.children = nil
		return true
	}

	// Otherwise the requested route diverges from the current node.
	nextBit := remainingRouteBits[0]
	if n.children[nextBit] == nil {
		// As an optimization, if we would create a node only to realize it is redundant, just
		// trim the rundundant child now and return.
		if remainingRouteBitsLen == 1 &&
			n.children[^nextBit&1] != nil &&
			n.children[^nextBit&1].isLeaf() {
			n.children = nil
			return true
		}

		// Otherwise we add a node
		n.children[nextBit] = new(node)
		if remainingRouteBitsLen > 1 {
			n.children[nextBit].children = new([2]*node)
		}
	}

	// Traverse to the new node
	if n.children[nextBit].insertRoute(remainingRouteBits[1:]) {
		return n.maybeRemoveRedundantChildren()
	}

	return false
}

// A node's children are redundant if they, taken together, represent a complete subtree from the
// node's perspective. This situation can be represented more simply as the node having a nil
// children pointer.
func (n *node) maybeRemoveRedundantChildren() bool {
	if n.isLeaf() {
		return false
	}

	if n.children[0] == nil ||
		!n.children[0].isLeaf() ||
		n.children[1] == nil ||
		!n.children[1].isLeaf() {
		return false
	}

	n.children = nil
	return true
}

type traversalStep struct {
	n                  *node
	precedingRouteBits bitslice.BitSlice
}

// Contents returns the BitSlices contained in the RSTree.
func (t *RSTree) Contents() []bitslice.BitSlice {
	// If the tree is empty
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

		if step.n.isLeaf() {
			contents = append(contents, step.precedingRouteBits)
		} else {
			lenPrecedingRouteBits := len(step.precedingRouteBits)

			if step.n.children[1] != nil {
				highChildBits := make([]byte, lenPrecedingRouteBits+1)
				copy(highChildBits, step.precedingRouteBits)
				highChildBits[lenPrecedingRouteBits] = 1
				remainingSteps.PushFront(traversalStep{
					n:                  step.n.children[1],
					precedingRouteBits: highChildBits,
				})
			}

			if step.n.children[0] != nil {
				lowChildBits := make([]byte, lenPrecedingRouteBits+1)
				copy(lowChildBits, step.precedingRouteBits)
				lowChildBits[lenPrecedingRouteBits] = 0
				remainingSteps.PushFront(traversalStep{
					n:                  step.n.children[0],
					precedingRouteBits: lowChildBits,
				})
			}
		}
	}

	return contents
}

func (t *RSTree) visitAll(cb func(*node)) {
	// If the trie is empty
	if t.root == nil {
		return
	}

	// Otherwise
	remainingSteps := list.New()
	remainingSteps.PushFront(traversalStep{
		n:                  t.root,
		precedingRouteBits: bitslice.BitSlice{},
	})

	for remainingSteps.Len() > 0 {
		curNode := remainingSteps.Remove(remainingSteps.Front()).(traversalStep)

		// Act on this node
		cb(curNode.n)

		// Traverse the remainder of the nodes
		if !curNode.n.isLeaf() {
			lenPrecedingRouteBits := len(curNode.precedingRouteBits)

			if curNode.n.children[1] != nil {
				highChildBits := make([]byte, lenPrecedingRouteBits+1)
				copy(highChildBits, curNode.precedingRouteBits)
				highChildBits[lenPrecedingRouteBits] = 1
				remainingSteps.PushFront(traversalStep{
					n:                  curNode.n.children[1],
					precedingRouteBits: highChildBits,
				})
			}

			if curNode.n.children[0] != nil {
				lowChildBits := make([]byte, lenPrecedingRouteBits+1)
				copy(lowChildBits, curNode.precedingRouteBits)
				lowChildBits[lenPrecedingRouteBits] = 0
				remainingSteps.PushFront(traversalStep{
					n:                  curNode.n.children[0],
					precedingRouteBits: lowChildBits,
				})
			}
		}
	}
}

// MemUsage returns information about an RSTrie's current size in memory.
func (t *RSTree) MemUsage() (uint, uint, uintptr, uintptr) {
	var numInternalNodes, numLeafNodes uint
	var internalNodesTotalSize, leafNodesTotalSize uintptr

	tallyNode := func(n *node) {
		baseNodeSize := unsafe.Sizeof(node{}) //nolint: exhaustruct, gosec
		if n.isLeaf() {
			numLeafNodes++
			leafNodesTotalSize += baseNodeSize
			return
		}

		numInternalNodes++
		internalNodesTotalSize += baseNodeSize + unsafe.Sizeof([2]*node{}) //nolint: gosec
	}
	t.visitAll(tallyNode)

	return numInternalNodes, numLeafNodes, internalNodesTotalSize, leafNodesTotalSize
}
