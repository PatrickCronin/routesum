// Package rstrie provides a datatype that supports building a space-efficient summary of networks
// and IPs.
package rstrie

import (
	"container/list"
	"unsafe"

	"github.com/PatrickCronin/routesum/pkg/routesum/routetype"
)

type route[T any] interface {
	*routetype.V4 | *routetype.V6
	Bits() uint8
	CommonAncestor(T) T
	Contains(T) bool
	NthBit(uint8) uint8
	Size() uintptr
}

type node[T route[T]] struct {
	route    T
	children *[2]*node[T]
}

// RSTrie is a radix-like trie of radix 2 whose stored "words" are the binary representations of
// networks and IPs. An optimization rstrie makes over a generic radix tree is that since routes
// covered by other routes don't need to be stored, each node in the trie will have either 0 or 2
// children; never 1.
type RSTrie[T route[T]] struct {
	root *node[T]
}

// New returns an initialized RSTrie for use
func New[T route[T]]() *RSTrie[T] {
	return &RSTrie[T]{
		root: nil,
	}
}

// InsertRoute inserts a new Route into the trie. Each insert results in a space-optimized trie
// structure representing its contents. If a route being inserted is already covered by an existing
// route, it's simply ignored. If a route being inserted covers one or more routes already in the
// trie, those nodes are removed and replaced by the new route.
func (t *RSTrie[T]) InsertRoute(r T) {
	// If the trie has no root node, simply create one to store the new route
	if t.root == nil {
		t.root = &node[T]{
			route:    r,
			children: nil,
		}
		return
	}

	t.root.insertRoute(&t.root, r)
}

func (n *node[T]) isLeaf() bool {
	return n.children == nil
}

func (n *node[T]) insertRoute(parent **node[T], r T) bool {
	// Does the requested route contain the current node's route? If so, update the current
	// node.
	if r.Contains(n.route) {
		n.route = r
		n.children = nil
		return true
	}

	if n.route.Contains(r) {
		// Does the current node cover the requested route? If so, we're done.
		if n.isLeaf() {
			return false
		}

		// Otherwise, we traverse to the correct child.
		traversedChild := r.NthBit(n.route.Bits() + 1)
		if n.children[traversedChild].insertRoute(&n.children[traversedChild], r) {
			return n.maybeRemoveRedundantChildren()
		}

		return false
	}

	// Otherwise the requested route diverges from the current node. We'll need to split the current
	// node.

	// As an optimization, if the split would result in a new node whose children represent a
	// complete subtrie, we just update the current node, instead of allocating new nodes and
	// optimizing them away immediately after.
	commonRouteAncestor := n.route.CommonAncestor(r)
	if n.isLeaf() &&
		n.route.Bits() == r.Bits() &&
		commonRouteAncestor.Bits() == n.route.Bits()-1 {
		n.route = commonRouteAncestor
		return true
	}

	var lowerChild, upperChild *node[T]
	if r.NthBit(commonRouteAncestor.Bits()+1) == 0 {
		lowerChild = &node[T]{route: r}
		upperChild = &node[T]{route: n.route, children: n.children}
	} else {
		lowerChild = &node[T]{route: n.route, children: n.children}
		upperChild = &node[T]{route: r}
	}

	*parent = &node[T]{
		route:    commonRouteAncestor,
		children: &[2]*node[T]{lowerChild, upperChild},
	}

	return n.maybeRemoveRedundantChildren()
}

// A node's children are redundant if they, taken together, represent a complete subtrie from the
// node's perspective. This situation can be represented more simply as the node having a nil
// children pointer.
func (n *node[T]) maybeRemoveRedundantChildren() bool {
	if n.isLeaf() {
		return false
	}

	if !n.children[0].isLeaf() || !n.children[1].isLeaf() {
		return false
	}

	if n.children[0].route.Bits() != n.children[1].route.Bits() ||
		n.children[0].route.Bits() != n.route.Bits()+1 {
		return false
	}

	n.children = nil
	return true
}

func (t *RSTrie[T]) visitAll(cb func(*node[T])) {
	// If the trie is empty
	if t.root == nil {
		return
	}

	// Otherwise
	remainingSteps := list.New()
	remainingSteps.PushFront(t.root)

	for remainingSteps.Len() > 0 {
		curNode := remainingSteps.Remove(remainingSteps.Front()).(*node[T])

		// Act on this node
		cb(curNode)

		// Traverse the remainder of the nodes
		if !curNode.isLeaf() {
			remainingSteps.PushFront(curNode.children[1])
			remainingSteps.PushFront(curNode.children[0])
		}
	}
}

// Contents returns the routes contained in the RSTrie.
func (t *RSTrie[T]) Contents() []T {
	var contents []T

	t.visitAll(func(n *node[T]) {
		if n.isLeaf() {
			contents = append(contents, n.route)
		}
	})

	return contents
}

// MemUsage returns information about an RSTrie's current size in memory.
func (t *RSTrie[T]) MemUsage() (uint, uint, uintptr, uintptr) {
	var numInternalNodes, numLeafNodes uint
	var internalNodesTotalSize, leafNodesTotalSize uintptr

	baseNodeSize := unsafe.Sizeof(node[T]{}) //nolint:  gosec

	t.visitAll(func(n *node[T]) {
		nodeSize := baseNodeSize + n.route.Size()
		if n.isLeaf() {
			numLeafNodes++
			leafNodesTotalSize += nodeSize
			return
		}

		numInternalNodes++
		internalNodesTotalSize += nodeSize + unsafe.Sizeof([2]node[T]{}) //nolint: gosec
	})

	return numInternalNodes, numLeafNodes, internalNodesTotalSize, leafNodesTotalSize
}
