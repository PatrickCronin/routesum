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
func (t *RSTrie[T]) InsertRoute(r T) { //nolint: funlen
	// If the trie has no root node, simply create one to store the new route
	if t.root == nil {
		t.root = &node[T]{
			route:    r,
			children: nil,
		}
		return
	}

	// Otherwise, perform a non-recursive search of the trie's nodes for the best place to insert
	// the route, and do so.
	visited := make([]*node[T], 0, 128)
	var traversedChild uint8
	curNode := t.root

	for {
		// Does the requested prefix contain the current node's prefix? If so, update the current
		// node.
		if r.Contains(curNode.route) {
			curNode.route = r
			curNode.children = nil
		}

		if curNode.route.Contains(r) {
			// Does the current node cover the requested route? If so, we're done.
			if curNode.isLeaf() {
				return
			}

			// Otherwise, we traverse to the correct child.
			visited = append(visited, curNode)
			traversedChild = r.NthBit(curNode.route.Bits() + 1)
			curNode = curNode.children[traversedChild]

			continue
		}

		// Otherwise the requested route diverges from the current node. We'll need to split the
		// current node.

		// As an optimization, if the split would result in a new node whose children represent a
		// complete subtrie, we just update the current node, instead of allocating new nodes and
		// optimizing them away immediately after.
		commonRouteAncestor := curNode.route.CommonAncestor(r)
		if curNode.isLeaf() &&
			curNode.route.Bits() == r.Bits() &&
			commonRouteAncestor.Bits() == curNode.route.Bits()-1 {
			curNode.route = commonRouteAncestor
		} else {
			var lowerChild, upperChild *node[T]
			if r.NthBit(commonRouteAncestor.Bits()+1) == 0 {
				lowerChild = &node[T]{route: r}
				upperChild = &node[T]{route: curNode.route, children: curNode.children}
			} else {
				lowerChild = &node[T]{route: curNode.route, children: curNode.children}
				upperChild = &node[T]{route: r}
			}
			newNode := &node[T]{
				route:    commonRouteAncestor,
				children: &[2]*node[T]{lowerChild, upperChild},
			}

			visitedLen := len(visited)
			if visitedLen == 0 {
				t.root = newNode
			} else {
				visited[visitedLen-1].children[traversedChild] = newNode
			}
		}

		simplifyVisitedSubtries(visited)
		return
	}
}

func (n *node[T]) isLeaf() bool {
	return n.children == nil
}

func (n *node[T]) childrenAreCompleteSubtrie() bool {
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

	return true
}

// A completed subtrie is a node in the trie whose children when taken together represent the
// complete subtrie below the node. For example, if a node represented the route "00", and it had a
// child for "0" and a child for "1", the node would be representing the "000" and "001" routes. But
// that's the same as having a single node for "00".
// simplifyCompletedSubtries takes a stack of visited nodes and simplifies completed subtries as far
// down the stack as possible. If at any point in the stack we find a node representing an
// incomplete subtrie, we stop.
func simplifyVisitedSubtries[T route[T]](visited []*node[T]) {
	for i := len(visited) - 1; i >= 0; i-- {
		if visited[i].isLeaf() {
			return
		}

		if !visited[i].childrenAreCompleteSubtrie() {
			return
		}

		visited[i].children = nil
	}
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
