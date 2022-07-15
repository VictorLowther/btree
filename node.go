package btree

// node[T] is a generic type that represents a node in the AVL tree.
type node[T any] struct {
	p *node[T] // parent
	l *node[T] // left child
	r *node[T] // right child
	h uint     // height of the node.
	i T        // The item the node is holding.
}

// balance calculates the relative balance of a node.
// Negative numbers indicate a subtree that is left-heavy,
// and positive numbers indicate a tree that is right-heavy.
func (n *node[T]) balance() (res int) {
	if n.l != nil {
		res -= int(n.l.h)
	}
	if n.r != nil {
		res += int(n.r.h)
	}
	return
}

// setHeight calculates the height of this node.
func (n *node[T]) setHeight() {
	n.h = 0
	if n.l != nil {
		n.h = n.l.h
	}
	if n.r != nil && n.r.h >= n.h {
		n.h = n.r.h
	}
	n.h++
	return
}

func (t *Tree[T]) newNode(v T) *node[T] {
	res := t.nodePool.Get().(*node[T])
	res.i = v
	res.h = 1
	t.count++
	t.insertCount++
	return res
}

func (t *Tree[T]) putNode(n *node[T]) {
	n.l = nil
	n.r = nil
	n.p = nil
	var ref T
	n.i = ref
	n.h = 0
	t.count--
	t.removeCount++
	t.nodePool.Put(n)
}

func (t *Tree[T]) copyNodes(n *node[T], into *Tree[T]) *node[T] {
	if n == nil {
		return nil
	}
	res := into.newNode(n.i)
	res.h = n.h
	if res.l = t.copyNodes(n.l, into); res.l != nil {
		res.l.p = res
	}
	if res.r = t.copyNodes(n.r, into); res.r != nil {
		res.r.p = res
	}
	return res
}

func (t *Tree[T]) releaseNodes(n *node[T]) {
	var s *node[T]
	for n != nil {
		if n.l != nil {
			n = n.l
			continue
		}
		if n.r != nil {
			n = n.r
			continue
		}
		if n.p != nil {
			n.p.swapChild(n, s)
		}
		s = n.p
		t.putNode(n)
		n = s
		s = nil
	}
}

// reverse reverses a tree by recursively swapping
// left and right pointers for all nodes.  This
// cannot break the AVL properties, so no other fields
// need to be adjusted.
func reverse[T any](n *node[T]) {
	if n.l != nil {
		reverse(n.l)
	}
	if n.r != nil {
		reverse(n.r)
	}
	n.l, n.r = n.r, n.l
}

func (n *node[T]) swapChild(was, is *node[T]) {
	if n.r == was {
		n.r = is
	} else {
		n.l = is
	}
	if is != nil {
		is.p = n
	}
}

// rotateLeft transforms
//
//   |
//   a
//  / \
// x   b
//    / \
//   y   z
//
// to
//     |
//     b
//    / \
//   a   z
//  / \
// x   y
func (a *node[T]) rotateLeft() (b *node[T]) {
	b = a.r
	if a.p != nil {
		a.p.swapChild(a, b)
	} else {
		b.p = nil
	}
	a.p = b
	if a.r = b.l; a.r != nil {
		a.r.p = a
	}
	b.l = a
	return
}

// rotateRight is the inverse of rotateLeft. it transforms
//
//     |
//     a(h)
//    / \
//   b   z
//  / \
// x   y
//
// to
//
//   |
//   b
//  / \
// x   a
//    / \
//   y   z
func (a *node[T]) rotateRight() (b *node[T]) {
	b = a.l
	if a.p != nil {
		a.p.swapChild(a, b)
	} else {
		b.p = nil
	}
	a.p = b
	if a.l = b.r; a.l != nil {
		a.l.p = a
	}
	b.r = a
	return
}

func (t *Tree[T]) getExact(n *node[T], v T) (res *node[T], found, onRight bool) {
	for n != nil {
		if t.less(v, n.i) {
			if n.l == nil {
				return n, false, false
			}
			n = n.l
		} else if t.less(n.i, v) {
			if n.r == nil {
				return n, false, true
			}
			n = n.r
		} else {
			break
		}
	}
	return n, true, false
}

// min finds the minimal child of h
func min[T any](n *node[T]) *node[T] {
	for n.l != nil {
		n = n.l
	}
	return n
}

// max finds the maximal child of h
func max[T any](n *node[T]) *node[T] {
	for n.r != nil {
		n = n.r
	}
	return n
}

// rebalanceAt walks up the tree starting at node n, rebalancing nodes
// that no longer meet the AVL balance criteria. rebalanceAt will continue until
// it either walks all the way up the tree, or the node has the
// same height it started with.
func (t *Tree[T]) rebalanceAt(n *node[T], forInsert bool) {
	for {
		oh := n.h
		switch n.balance() {
		case Less, Equal, Greater:
		case 2:
			// Tree is excessively right-heavy, rotate it to the left.
			if n.r != nil && n.r.balance() < 0 {
				// Right tree is left-heavy, which would cause the next rotation to result in overall left-heaviness.
				// Rotate the right tree to the right to counteract this.
				n.r = n.r.rotateRight()
				n.r.r.setHeight()
			}
			n = n.rotateLeft()
			n.l.setHeight()
			if forInsert {
				t.insertRebalanceCount++
			} else {
				t.removeRebalanceCount++
			}
		case -2:
			// Tree is excessively left-heavy, rotate it to the right
			if n.l != nil && n.l.balance() > 0 {
				// The left tree is right-heavy, which would cause the next rotation to result in overall right-heaviness.
				// Rotate the left tree to the left to compensate.
				n.l = n.l.rotateLeft()
				n.l.l.setHeight()
			}
			n = n.rotateRight()
			n.r.setHeight()
			if forInsert {
				t.insertRebalanceCount++
			} else {
				t.removeRebalanceCount++
			}
		default:
			panic("Tree too far out of shape!")
		}
		n.setHeight()
		if n.p == nil {
			t.root = n
			return
		}
		if oh == n.h {
			return
		}
		n = n.p
	}
}

// insert or replace a new value. If a new value is inserted, any needed rebalancing
// is performed.
func (t *Tree[T]) insert(v T) {
	n, found, onRight := t.getExact(t.root, v)
	if found {
		n.i = v
		return
	}
	nn := t.newNode(v)
	nn.p = n
	if onRight {
		n.r = nn
		onRight = n.l == nil
	} else {
		n.l = nn
		onRight = n.r == nil
	}

	if onRight {
		n.h++
		if n.p != nil {
			t.rebalanceAt(n.p, true)
		}
	}
}

// remove the passed-in value from the tree, if it exists. The tree will be rebalanced if needed.
func (t *Tree[T]) remove(v T) (deleted T, found bool) {
	var (
		at, alt *node[T]
	)
	at, found, _ = t.getExact(t.root, v)
	if !found {
		return
	}
	deleted = at.i
	for {
		if at.h == 1 {
			if alt = at.p; alt != nil {
				alt.swapChild(at, at.r)
				t.rebalanceAt(alt, false)
			} else {
				t.root = nil
			}
			t.putNode(at)
			return
		} else if at.r != nil {
			alt = min(at.r)
		} else if at.l != nil {
			alt = max(at.l)
		} else {
			panic("Impossible")
		}
		at.i, alt.i = alt.i, at.i
		at = alt
	}
}
