package btree

// Test is a function signature that is used for iterating through
// a tree along with the signature that Range, Before, and After
// discriminators must match.
type Test[T any] func(T) bool

// TestMaker is a function that takes a CompareAgainst and makes a Test from it.
type TestMaker[T any] func(CompareAgainst[T]) Test[T]

// Lt is a TestMaker that returns true if the item in the
// tree being examined is less than the item the CompareAgainst function wraps.
func Lt[T any](c CompareAgainst[T]) Test[T] {
	return func(idx T) bool { return c(idx) == Less }
}

// Lte is a TestMaker that returns true if the item in the
// tree being examined is less than or equal to the item the CompareAgainst function wraps.
func Lte[T any](c CompareAgainst[T]) Test[T] {
	return func(idx T) bool { return c(idx) < Greater }
}

// Eq is a TestMaker that returns true if the item in the
// tree being examined is equal to the item the CompareAgainst function wraps.
func Eq[T any](c CompareAgainst[T]) Test[T] {
	return func(idx T) bool { return c(idx) == Equal }
}

// Gte is a TestMaker that returns true if the item in the
// tree being examined is greater than or equal to the item the CompareAgainst function wraps.
func Gte[T any](c CompareAgainst[T]) Test[T] {
	return func(idx T) bool { return c(idx) > Less }
}

// Gt is a TestMaker that returns true if the item in the
// tree being examined is greater than the item the CompareAgainst function wraps.
func Gt[T any](c CompareAgainst[T]) Test[T] {
	return func(idx T) bool { return c(idx) == Greater }
}

// Ne is a TestMaker that returns true if the item in the
// tree being examined is not equal to the item the CompareAgainst function wraps.
func Ne[T any](c CompareAgainst[T]) Test[T] {
	return func(idx T) bool { return c(idx) != Equal }
}

// Iterator holds state needed to iterate over a binary tree.
// You must not modify the tree while iterating over it, lest you
// get undefined results and/or panics.
type Iterator[T any] struct {
	t           *Tree[T]
	stack       []*node[T]
	workingNode *node[T]
	start, stop Test[T]
	ascending   bool
}

func (i *Iterator[T]) clearStack() {
	for k := range i.stack {
		i.stack[k] = nil
	}
	i.stack = i.stack[:0]
}

// Release releases the state the Iterator holds.
// Subsequent calls to Next will return false, and subsequent
// calls to Item will panic.
func (i *Iterator[T]) Release() {
	i.clearStack()
	i.workingNode = nil
	i.start = nil
	i.stop = nil
	i.t = nil
}

func (i *Iterator[T]) stackHead() *node[T] {
	switch idx := len(i.stack); idx {
	case 0:
		return nil
	default:
		return i.stack[idx-1]
	}
}

func (i *Iterator[T]) push(n *node[T]) {
	i.stack = append(i.stack, n)
}

func (i *Iterator[T]) pop() {
	switch idx := len(i.stack); idx {
	case 0:
		i.workingNode = nil
		return
	default:
		tos := idx - 1
		i.stack[tos] = nil
		i.stack = i.stack[:tos]
		if tos > 0 {
			i.workingNode = i.stack[tos-1]
		} else {
			i.workingNode = nil
		}
	}
	return
}

func (i *Iterator[T]) swapHead() {
	i.stack[len(i.stack)-1] = i.workingNode
}

// Item returns the item that the current node points to.
// It will panic if iteration has not yet started, or if iteration has finished.
func (i *Iterator[T]) Item() T {
	if len(i.stack) == 0 {
		panic("No iteration in progress")
	}
	return i.workingNode.i
}

func (i *Iterator[T]) pickNextNode(current, next, bound *node[T], boundCheck Test[T]) *node[T] {
	if boundCheck != nil && boundCheck(current.i) {
		return bound
	}
	i.push(current)
	return next
}

func (i *Iterator[T]) min(n *node[T]) {
	for n != nil {
		n = i.pickNextNode(n, n.l, n.r, i.start)
	}
}

func (i *Iterator[T]) max(n *node[T]) {
	for n != nil {
		n = i.pickNextNode(n, n.r, n.l, i.stop)
	}
}

func (i *Iterator[T]) init(ascending bool, orNot Test[T]) bool {
	if i.workingNode != nil {
		if ascending {
			i.min(i.workingNode)
		} else {
			i.max(i.workingNode)
		}
		i.ascending = ascending
		i.workingNode = i.stackHead()
		if i.workingNode == nil || (orNot != nil && orNot(i.workingNode.i)) {
			i.Release()
			return false
		}
		return true
	} else {
		i.Release()
		return false
	}
}

func (i *Iterator[T]) changeDirection() bool {
	i.ascending = !i.ascending
	i.clearStack()
	v := i.workingNode.i
	i.workingNode = i.t.root
	var old Test[T]
	if i.ascending {
		old = i.start
		i.start = Lte(i.t.Cmp(v))
		if !i.Next() {
			return false
		}
		i.start = old
	} else {
		old = i.stop
		i.stop = Gte(i.t.Cmp(v))
		if !i.Prev() {
			return false
		}
		i.stop = old
	}
	return true
}

// Next walks to the next larger node in the tree and returns true,
// or returns false if there is no next larger node to walk to.
//
// If Next returns true, Item will return the item that
// the current node contains.
func (i *Iterator[T]) Next() bool {
	if len(i.stack) == 0 {
		return i.init(true, i.stop)
	}
	if !i.ascending && !i.changeDirection() {
		return false
	}
	if i.workingNode.r == nil {
		i.pop()
	} else {
		i.workingNode = i.workingNode.r
		i.swapHead()
		if i.workingNode.l != nil {
			i.min(i.workingNode.l)
			i.workingNode = i.stackHead()
		}
	}
	if i.workingNode == nil || (i.stop != nil && i.stop(i.workingNode.i)) {
		i.Release()
		return false
	}
	return true
}

// Prev walks to the next smaller node in the tree and returns true,
// or returns false if there is no next smaller node to walk to.
//
// If Prev returns true, Item will return the item that
// the current node contains.
func (i *Iterator[T]) Prev() bool {
	if len(i.stack) == 0 {
		return i.init(false, i.stop)
	}
	if i.ascending && !i.changeDirection() {
		return false
	}
	if i.workingNode.l == nil {
		i.pop()
	} else {
		i.workingNode = i.workingNode.l
		i.swapHead()
		if i.workingNode.r != nil {
			i.max(i.workingNode.r)
			i.workingNode = i.stackHead()
		}
	}
	if i.workingNode == nil || (i.start != nil && i.start(i.workingNode.i)) {
		i.Release()
		return false
	}
	return true
}

// Iterator creates a new Iterator that will ignore all items on the left for which start returns true and
// all items on the right for which stop returns true.
//
// start should be one of Lt (inclusive), Lte (exclusive)
//
// stop should be one of Gt (inclusive), Gte (exclusive)
//
// If either start or stop is nil, then that condition will not apply.
//
// After the Iterator is created, you must call Next or Prev to
// fetch the initial item.  Calling Next will start iteration with the
// smallest item in the tree that start returns false for, and calling Prev
// will start iteration with the largest item in the tree that stop returns
// false for.
//
// Example:
//
//    iter := tree.Iterator(nil,nil)
//    for iter.Next() {
//        fmt.Println(iter.Item())
//    }
//
// will print all the items in tree in order.
func (t *Tree[T]) Iterator(start, stop Test[T]) *Iterator[T] {
	return &Iterator[T]{
		t:           t,
		workingNode: t.root,
		start:       start,
		stop:        stop,
	}
}

// Range will iterate through the tree in ascending order,
// ignoring all items to the left that start returns true for
// and all items in the right that end returns true for.
// Iteration will also stop if iterator returns false.
//
// Lt  start == inclusive, Lte start == exclusive
// Gte stop  == exclusive, Gt  stop  == inclusive
func (t *Tree[T]) Range(start, stop, iterator Test[T]) {
	i := t.Iterator(start, stop)
	for i.Next() {
		if !iterator(i.Item()) {
			i.Release()
		}
	}
}

// After will iterate through the tree in ascending order
// ignoring items on the left that start returns true for.
// Iteration will also stop when iterator returns false.
//
// Lt start == inclusive, Lte start = exclusive
func (t *Tree[T]) After(start, iterator Test[T]) {
	i := t.Iterator(start, nil)
	for i.Next() {
		if !iterator(i.Item()) {
			i.Release()
		}
	}
}

// Before will iterate through the tree in ascending order
// ignoring items on the right that end returns true for.
// Iteration will stop if iterator returns false.
//
// Gt stop == inclusive, Gte stop = exclusive
func (t *Tree[T]) Before(stop, iterator Test[T]) {
	i := t.Iterator(nil, stop)
	for i.Next() {
		if !iterator(i.Item()) {
			i.Release()
		}
	}
}

// Walk will call Iterator once for each item in the tree in ascending order.
// Walk will return early if iterator returns false.
func (t *Tree[T]) Walk(iterator Test[T]) {
	i := t.Iterator(nil, nil)
	for i.Next() {
		if !iterator(i.Item()) {
			i.Release()
		}
	}
}
