package btree

import "sync"

const (
	Less    = -1
	Equal   = 0
	Greater = 1
)

// CompareAgainst is a comparison function that compares a reference item to
// an item in the Tree.
// An example of how it should work:
//
//    comparer := func(reference i) CompareAgainst {
//        return func(treeItem i) int {
//            switch {
//            case LessThan(treeItem, reference): return Less
//            case LessThan(reference, treeItem): return Greater
//            default: return Equal
//            }
//        }
//    }
//
// CompareAgainst must return:
//
// * Less if the item in the tree is less than the reference
//
// * Equal if the item in the tree is equal to the reference
//
// * Greater if the item in the tree is greater than the reference
type CompareAgainst[T any] func(T) int

// LessThan compares two values to see if the first is LessThan than
// the second.  The Tree code considers any values where neither is LessThan the other
// to be equal.
type LessThan[T any] func(T, T) bool

// Tree is an AVL tree.
type Tree[T any] struct {
	root                              *node[T]
	less                              LessThan[T]
	nodePool                          *sync.Pool
	insertCount, insertRebalanceCount uint64
	removeCount, removeRebalanceCount uint64
	count                             int
}

// New allocates a new Tree that will keep itself ordered according to the passed in LessThan.
func New[T any](lt LessThan[T]) *Tree[T] {
	res := &Tree[T]{}
	res.less = lt
	res.nodePool = &sync.Pool{New: func() any { return &node[T]{} }}
	return res
}

// Cmp takes a reference T and makes a valid CompareAgainst
// using the tree's current LessThan comparator.
func (t *Tree[T]) Cmp(reference T) CompareAgainst[T] {
	less := t.less
	return func(treeVal T) int {
		if less(treeVal, reference) {
			return Less
		}
		if less(reference, treeVal) {
			return Greater
		}
		return Equal
	}
}

// Release caches the memory that the Tree refers to for later reuse.
// You must not reuse any part of the Tree after calling Release.
func (t *Tree[T]) Release() {
	t.count = 0
	t.less = nil
	if t.root != nil {
		t.releaseNodes(t.root)
		t.root = nil
	}
}

// Reverse reverses a Tree in-place by swizzling the pointers in the nodes
// around and inverting the ordering function. This avoids needing to
// make a copy of the tree and resort the data.  If you want to do that,
// make a Clone of the Tree and Reverse that.
func (t *Tree[T]) Reverse() {
	ll := t.less
	t.less = func(a, b T) bool { return ll(b, a) }
	if t.root == nil {
		return
	}
	var n *node[T]
	i := t.Iterator(nil, nil)
	for i.Next() {
		if n != nil {
			n.r, n.l = n.l, n.r
		}
		n = i.workingNode
	}
	n.r, n.l = n.l, n.r
}

// Copy makes a new copy of the Tree that has the same ordering function
// but no data.  Trees created using Copy (or any functions that use it)
// use the same sync.Pool of nodes.
func (t *Tree[T]) Copy() *Tree[T] {
	res := New[T](t.less)
	res.nodePool = t.nodePool
	return res
}

// Clone makes a full copy of the Tree, including all data.
func (t *Tree[T]) Clone() *Tree[T] {
	res := t.Copy()
	res.root = t.copyNodes(t.root, res)
	return res
}

// SortBy returns a new empty Tree with an ordering function that falls back to
// t.less if the passed-in LessThan considers two items to be equal.
// This (and SortedClone) can be used to implement trees that will maintain items in
// arbitrarily complicated sort orders.
func (t *Tree[T]) SortBy(l LessThan[T]) *Tree[T] {
	prevLess := t.less
	res := New[T](func(a, b T) bool {
		switch {
		case l(a, b):
			return true
		case l(b, a):
			return false
		default:
			return prevLess(a, b)
		}
	})
	res.nodePool = t.nodePool
	return res
}

// SortedClone makes a new Tree using SortBy, then inserts all the data from t into it.
func (t *Tree[T]) SortedClone(l LessThan[T]) *Tree[T] {
	res := t.SortBy(l)
	iter := t.Iterator(nil, nil)
	for iter.Next() {
		res.Insert(iter.Item())
	}
	return res
}

// Len returns the number of nodes in the tree.
func (t *Tree[T]) Len() int { return t.count }

const unorderable = `Unorderable CompareAgainst passed to Get`

// Get returns either the highest item in the tree that is equal to CompareAgainst and true,
// or a zero T and false if there is no such value in the Tree.
// The Tree must be sorted at the top level in the order that CompareAgainst expects, or you
// will get nonsense results.  If you want to retrieve all
// the items matching CompareAgainst, use one of the Range, Before, or After instead.
func (t *Tree[T]) Get(cmp CompareAgainst[T]) (item T, found bool) {
	h := t.root
	for h != nil {
		switch cmp(h.i) {
		case Greater:
			h = h.l
		case Less:
			h = h.r
		case Equal:
			item, found = h.i, true
			return
		default:
			panic(unorderable)
		}
	}
	return
}

// Has returns true if the tree contains an element equal to CompareAgainst.
func (t *Tree[T]) Has(cmp CompareAgainst[T]) bool {
	_, found := t.Get(cmp)
	return found
}

// Min returns the smallest item in the Tree and true, or a zero T and false if the tree is empty.
func (t *Tree[T]) Min() (item T, found bool) {
	if t.root != nil {
		found = true
		item = min(t.root).i
	}
	return
}

// Max returns the largest item in the Tree and true, or a zero T and false if the tree is empty.
func (t *Tree[T]) Max() (item T, found bool) {
	if t.root != nil {
		found = true
		item = max(t.root).i
	}
	return
}

// Insert an item into the tree. If there is an existing value in the Tree that
// is equal to item, item will replace that value.
func (t *Tree[T]) Insert(item T) {
	if t.root == nil {
		t.root = t.newNode(item)
	} else {
		t.insert(item)
	}
}

// Delete item from the tree, returning the item deleted
// or an empty i if the item was not in the tree.
func (t *Tree[T]) Delete(item T) (deleted T, found bool) {
	if t.root == nil {
		return
	}
	deleted, found = t.remove(item)
	return
}
