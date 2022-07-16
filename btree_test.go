package btree

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

func (n *node[T]) height() uint {
	if n == nil {
		return 0
	}
	return n.h
}

var intPool = &sync.Pool{New: func() any { return &node[int]{} }}
var stringPool = &sync.Pool{New: func() any { return &node[string]{} }}

// balanced checks a tree to ensure it is AVL compliant.
// Only for use when running tests.
func (n *node[T]) balanced(t *testing.T) {
	if n == nil {
		return
	}
	if n.h == 0 {
		panic("Zero height")
	}
	if n.h == 1 && !(n.r == nil && n.l == nil) {
		panic("Height 1 node has children")
	}
	if n.h > 1 && n.r == nil && n.l == nil {
		panic("Interior node has no children")
	}
	lh, rh := n.l.height(), n.r.height()
	if lh >= n.h || rh >= n.h {
		panic("Child height greater than ours")
	}
	if !(n.h-lh == 1 || n.h-rh == 1) {
		panic("Height not max(lh,rh)+1")
	}
	b := n.balance()
	rb := int(rh) - int(lh)
	if b != rb {
		panic("Balance calculated incorrectly")
	}
	if b > 1 {
		panic("Too heavy to the right!")
	} else if b < -1 {
		panic("Too heavy to the left!")
	}
	if n.r != nil && n.r.p != n {
		panic("Right parent not set correctly")
	}
	if n.l != nil && n.l.p != n {
		panic("Less parent not set correctly")
	}
	if n.l == n || n.r == n || n.p == n {
		panic("Fatal self recursion")
	}
	if n.l != nil {
		n.l.balanced(t)
	}
	if n.r != nil {
		n.r.balanced(t)
	}
}

func newIntTree() (*Tree[int], func(int) CompareAgainst[int]) {
	tree := New[int](func(a, b int) bool { return a < b })
	tree.nodePool = intPool
	return tree,
		func(v int) CompareAgainst[int] {
			return func(vv int) int {
				switch {
				case vv < v:
					return -1
				case vv > v:
					return 1
				default:
					return 0
				}
			}
		}
}

func newStringTree() (*Tree[string], func(string) CompareAgainst[string]) {
	tree := New[string](func(a, b string) bool { return a < b })
	tree.nodePool = stringPool
	return tree,
		func(v string) CompareAgainst[string] {
			return func(idx string) int {
				switch {
				case idx < v:
					return -1
				case idx > v:
					return 1
				default:
					return 0
				}
			}
		}
}

func TestRotate(t *testing.T) {
	tree := New[int](func(a, b int) bool { return a < b })
	tree.Insert(1)
	tree.Insert(0)
	tree.Insert(3)
	tree.Insert(2)
	tree.Insert(4)
	if tree.root.i != 1 {
		t.Fatalf("tree.root.i %d, not 1", tree.root.i)
	}
	if tree.root.l.i != 0 {
		t.Fatalf("tree root.l.i %d, not 0", tree.root.l.i)
	}
	if tree.root.r.i != 3 {
		t.Fatalf("tree.root.r.i %d, not 3", tree.root.r.i)
	}
	if tree.root.r.l.i != 2 {
		t.Fatalf("tree.root.r.l.i %d, not 2", tree.root.r.l.i)
	}
	if tree.root.r.r.i != 4 {
		t.Fatalf("tree.root.r.r.i %d, not 4", tree.root.r.r.i)
	}
	t.Logf("Tree populated correctly")
	tree.root.balanced(t)
	tree.root = tree.root.rotateLeft()
	tree.root.l.setHeight()
	tree.root.setHeight()
	if tree.root.i != 3 {
		t.Fatalf("tree.root.i %d, not 3", tree.root.i)
	}
	if tree.root.l.i != 1 {
		t.Fatalf("tree root.l.i %d, not 1", tree.root.l.i)
	}
	if tree.root.l.l.i != 0 {
		t.Fatalf("tree.root.l.l.i %d, not 0", tree.root.l.l.i)
	}
	if tree.root.l.r.i != 2 {
		t.Fatalf("tree.root.l.r.i %d, not 2", tree.root.l.r.i)
	}
	if tree.root.r.i != 4 {
		t.Fatalf("tree.root.r.i %d, not 4", tree.root.r.i)
	}
	t.Logf("Tree rotated left correctly")
	tree.root.balanced(t)
	tree.root = tree.root.rotateRight()
	tree.root.r.setHeight()
	tree.root.setHeight()
	if tree.root.i != 1 {
		t.Fatalf("tree.root.i %d, not 1", tree.root.i)
	}
	if tree.root.l.i != 0 {
		t.Fatalf("tree root.l.i %d, not 0", tree.root.l.i)
	}
	if tree.root.r.i != 3 {
		t.Fatalf("tree.root.r.i %d, not 3", tree.root.r.i)
	}
	if tree.root.r.l.i != 2 {
		t.Fatalf("tree.root.r.l.i %d, not 2", tree.root.r.l.i)
	}
	if tree.root.r.r.i != 4 {
		t.Fatalf("tree.root.r.r.i %d, not 4", tree.root.r.r.i)
	}
	tree.root.balanced(t)
	t.Logf("Tree rotated right correctly")
	tree.Reverse()
	tree.root.balanced(t)
}

func TestCases(t *testing.T) {
	tree, cmp := newIntTree()
	defer tree.Release()
	tree.Insert(1)

	if tree.Len() != 1 {
		t.Fatalf("expecting len 1")
	}
	if !tree.Has(cmp(1)) {
		t.Fatalf("expecting to find key=1")
	}

	tree.Delete(1)
	if tree.Len() != 0 {
		t.Fatalf("expecting len 0")
	}
	if tree.Has(cmp(1)) {
		t.Fatalf("not expecting to find key=1")
	}

	tree.Delete(1)
	if tree.Len() != 0 {
		t.Fatalf("expecting len 0")
	}
	if tree.Has(cmp(1)) {
		t.Fatalf("not expecting to find key=1")
	}
}

func TestRange(t *testing.T) {
	tree, cmp := newStringTree()
	defer tree.Release()
	for _, v := range []string{"ab", "aba", "abc", "a", "aa", "aaa", "b", "a-", "a!"} {
		tree.Insert(v)
	}
	expect := []string{"ab", "aba", "abc"}
	res := []string{}
	tree.Range(Lt(cmp("ab")), Gt(cmp("ac")), func(idx string) bool {
		res = append(res, idx)
		return true
	})
	if !reflect.DeepEqual(expect, res) {
		t.Fatalf("Range failed: expected %v, got %v", expect, res)
	}
	res = nil
	tree.Range(Lte(cmp("aaa")), Gte(cmp("b")), func(idx string) bool {
		res = append(res, idx)
		return true
	})
	if !reflect.DeepEqual(expect, res) {
		t.Fatalf("Range failed: expected %v, got %v", expect, res)
	}
}

func TestIter(t *testing.T) {
	tree, cmp := newStringTree()
	defer tree.Release()
	for _, v := range []string{"ab", "aba", "abc", "a", "aa", "aaa", "b", "a-", "a!"} {
		tree.Insert(v)
	}
	expect := []string{"ab", "aba", "abc"}
	res := []string{}
	iter := tree.Iterator(Lt(cmp("ab")), Gt(cmp("ac")))
	for iter.Next() {
		res = append(res, iter.Item())
	}
	if !reflect.DeepEqual(expect, res) {
		t.Fatalf("Range failed: expected %v, got %v", expect, res)
	}
	res = nil
	iter = tree.Iterator(Lte(cmp("aaa")), Gte(cmp("b")))
	for iter.Next() {
		res = append(res, iter.Item())
	}
	if !reflect.DeepEqual(expect, res) {
		t.Fatalf("Range failed: expected %v, got %v", expect, res)
	}
	res = nil
	expect = nil
	iter = tree.Iterator(Lt(cmp("z")), nil)
	for iter.Next() {
		res = append(res, iter.Item())
	}
	if !reflect.DeepEqual(expect, res) {
		t.Fatalf("Range failed: expected %v, got %v", expect, res)
	}
	iter = tree.Iterator(nil, Gt(cmp("0")))
	for iter.Next() {
		res = append(res, iter.Item())
	}
	if !reflect.DeepEqual(expect, res) {
		t.Fatalf("Range failed: expected %v, got %v", expect, res)
	}
}

func TestIterDirection(t *testing.T) {
	tree, cmp := newIntTree()
	for i := 0; i < 100; i++ {
		tree.Insert(i)
	}
	for _, idx := range []int{0, 10, 90} {
		iter := tree.Iterator(Lt(cmp(idx)), nil)
		i := idx
		for iter.Next() {
			if iter.Item() != i {
				t.Fatalf("%d != %d", iter.Item(), i)
			}
			i++
		}
		iter = tree.Iterator(nil, Gt(cmp(idx)))
		i = idx
		for iter.Prev() {
			if iter.Item() != i {
				t.Fatalf("%d != %d", iter.Item(), i)
			}
			i--
		}
	}
}

func TestReverse(t *testing.T) {
	tree, _ := newIntTree()
	defer tree.Release()
	src := rand.New(rand.NewSource(55))
	n := 1000
	for _, v := range src.Perm(n) {
		tree.Insert(v)
		tree.root.balanced(t)
	}
	j := 0
	iter := tree.Iterator(nil, nil)
	for iter.Next() {
		if iter.Item() != j {
			t.Fatalf("bad order")
		}
		j++
	}
	tree.Reverse()
	j = n
	iter = tree.Iterator(nil, nil)
	for iter.Next() {
		j--
		if iter.Item() != j {
			t.Fatalf("bad order")
		}
	}
}

func TestRandomInsertOrder(t *testing.T) {
	tree, _ := newIntTree()
	defer tree.Release()
	src := rand.New(rand.NewSource(0))
	n := 10000
	for _, v := range src.Perm(n) {
		tree.Insert(v)
		tree.root.balanced(t)
	}
	j := 0
	tree.Walk(func(idx int) bool {
		if idx != j {
			t.Fatalf("bad order")
		}
		j++
		return true
	})
}

func TestRandomInsertDelete(t *testing.T) {
	tree, _ := newIntTree()
	defer tree.Release()
	n := 10000
	src := rand.New(rand.NewSource(0))
	backing := src.Perm(n)
	for i := 0; i < n; i++ {
		tree.Insert(backing[i])
		tree.root.balanced(t)
	}
	for i := 0; i < n; i++ {
		idx, found := tree.Delete(backing[i])
		tree.root.balanced(t)
		if !found {
			t.Fatalf("Did not find %d in the tree at %d", backing[i], i)
		}
		if idx != backing[i] {
			t.Fatalf("Error deleting: wanted %d, got %d", backing[i], idx)
		}
	}
	ins, rms, rbi, rbr := tree.RebalanceStats()
	t.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func TestRandomInsertDeleteNonExistent(t *testing.T) {
	tree, _ := newIntTree()
	defer tree.Release()
	n := 100
	backing := rand.Perm(n)
	for i := 0; i < n; i++ {
		tree.Insert(backing[i])
		tree.root.balanced(t)
	}
	if v, found := tree.Delete(200); found {
		t.Fatalf("deleted non-existent item %d", v)
	}
	if v, found := tree.Delete(-2); found {
		t.Fatalf("deleted non-existent item %d", v)
	}
	for i := 0; i < n; i++ {
		if _, found := tree.Delete(i); !found {
			t.Fatalf("remove failed for %d", i)
		}
		tree.root.balanced(t)
	}
	if v, found := tree.Delete(200); found {
		t.Fatalf("deleted non-existent item %d", v)
	}
	if v, found := tree.Delete(-2); found {
		t.Fatalf("deleted non-existent item %d", v)
	}
	if tree.Len() != 0 {
		t.Fatalf("Failed to remove %d items!", tree.Len())
	}
	ins, rms, rbi, rbr := tree.RebalanceStats()
	t.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func TestRandomInsertPartialDeleteOrder(t *testing.T) {
	tree, _ := newIntTree()
	defer tree.Release()
	n := 1000
	backing := rand.Perm(n)
	for i := 0; i < n; i++ {
		tree.Insert(backing[i])
		tree.root.balanced(t)
	}
	for i := 0; i < n; i++ {
		if _, found := tree.Delete(i); !found {
			t.Fatalf("remove failed")
		}
		tree.root.balanced(t)
	}
	if tree.Len() != 0 {
		t.Fatalf("Failed to remove %d items!", tree.Len())
	}
	ins, rms, rbi, rbr := tree.RebalanceStats()
	t.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func TestRandomInsertStats(t *testing.T) {
	tree, _ := newIntTree()
	defer tree.Release()
	n := 100000
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for _, i := range r.Perm(n) {
		tree.Insert(i)
	}
	avg, _ := tree.HeightStats()
	expAvg := math.Log2(float64(n)) - 1.5
	if math.Abs(avg-expAvg) >= 1.44 {
		t.Errorf("too much deviation from expected average height")
	}
	ins, rms, rbi, rbr := tree.RebalanceStats()
	t.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func TestSeqInsertStats(t *testing.T) {
	tree, _ := newIntTree()
	defer tree.Release()
	n := 100000
	for i := 0; i < n; i++ {
		tree.Insert(i)
	}
	avg, _ := tree.HeightStats()
	expAvg := math.Log2(float64(n)) - 1.5
	if math.Abs(avg-expAvg) >= 1.44 {
		t.Errorf("too much deviation from expected average height")
	}
	ins, rms, rbi, rbr := tree.RebalanceStats()
	t.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func BenchmarkInsertIntSeq(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	tree, _ := newIntTree()
	defer tree.Release()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tree.Insert(i)
	}
	b.StopTimer()
	//ins, rms, rbi, rbr := tree.RebalanceStats()
	//b.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func BenchmarkInsertIntSeqReverse(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	tree, _ := newIntTree()
	defer tree.Release()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tree.Insert(b.N - i)
	}
	b.StopTimer()
	//ins, rms, rbi, rbr := tree.RebalanceStats()
	//b.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func BenchmarkInsertIntRand(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	seed := time.Now().Unix()
	tree, _ := newIntTree()
	defer tree.Release()
	rs := rand.New(rand.NewSource(seed))
	backing := rs.Perm(b.N)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tree.Insert(backing[i])
	}
	b.StopTimer()
	//ins, rms, rbi, rbr := tree.RebalanceStats()
	//b.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func BenchmarkDeleteIntSeq(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	tree, _ := newIntTree()
	defer tree.Release()
	for i := 0; i < b.N; i++ {
		tree.Insert(i)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tree.Delete(i)
	}
	b.StopTimer()
	//ins, rms, rbi, rbr := tree.RebalanceStats()
	//b.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func BenchmarkDeleteIntRand(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	seed := time.Now().Unix()
	tree, _ := newIntTree()
	defer tree.Release()
	rs := rand.New(rand.NewSource(seed))
	vals := rs.Perm(b.N)
	for _, v := range vals {
		tree.Insert(v)
	}
	b.StartTimer()
	for _, v := range vals {
		tree.Delete(v)
	}
	b.StopTimer()
	//ins, rms, rbi, rbr := tree.RebalanceStats()
	//b.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func BenchmarkInsertStringSeq(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	seed := time.Now().Unix()
	tree, _ := newStringTree()
	defer tree.Release()
	rs := rand.New(rand.NewSource(seed))
	backing := make([]string, b.N)
	buf := [32]byte{}
	for i := range backing {
		rs.Read(buf[:])
		backing[i] = string(append([]byte{}, buf[:]...))
	}
	sort.Strings(backing)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tree.Insert(backing[i])
	}
	b.StopTimer()
}

func BenchmarkInsertStringRand(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	seed := time.Now().Unix()
	tree, _ := newStringTree()
	defer tree.Release()
	rs := rand.New(rand.NewSource(seed))
	backing := make([]string, b.N)
	buf := [32]byte{}
	for i := range backing {
		rs.Read(buf[:])
		backing[i] = string(append([]byte{}, buf[:]...))
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tree.Insert(backing[i])
	}
	b.StopTimer()
}

func BenchmarkDeleteStringSeq(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	seed := time.Now().Unix()
	tree, _ := newStringTree()
	defer tree.Release()
	rs := rand.New(rand.NewSource(seed))
	backing := make([]string, b.N)
	buf := [32]byte{}
	for i := range backing {
		rs.Read(buf[:])
		backing[i] = string(append([]byte{}, buf[:]...))
	}
	sort.Strings(backing)
	for _, v := range backing {
		tree.Insert(v)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tree.Delete(backing[i])
	}
	b.StopTimer()
	//ins, rms, rbi, rbr := tree.RebalanceStats()
	//b.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func BenchmarkDeleteStringRand(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	seed := time.Now().Unix()
	tree, _ := newStringTree()
	defer tree.Release()
	rs := rand.New(rand.NewSource(seed))
	backing := make([]string, b.N)
	buf := [32]byte{}
	for i := range backing {
		rs.Read(buf[:])
		backing[i] = string(append([]byte{}, buf[:]...))
	}
	for _, v := range backing {
		tree.Insert(v)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tree.Delete(backing[i])
	}
	b.StopTimer()
	//ins, rms, rbi, rbr := tree.RebalanceStats()
	//b.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func BenchmarkIntIterAll(b *testing.B) {
	b.Skipf("Long running and memory heavy")
	b.StopTimer()
	tree, _ := newIntTree()
	defer tree.Release()
	for i := 0; i < b.N; i++ {
		tree.Insert(i)
	}
	b.StartTimer()
	all := tree.Iterator(nil, nil)
	i := 0
	for all.Next() {
		if i != all.Item() {
			b.Fatal(i, " != ", all.Item())
		}
		i++
	}
	b.StopTimer()
}

func BenchmarkIntIterAfter(b *testing.B) {
	b.Skipf("Long running and memory heavy")
	b.StopTimer()
	tree, cmp := newIntTree()
	defer tree.Release()
	for i := 0; i < b.N; i++ {
		tree.Insert(i)
	}
	b.StartTimer()
	i := b.N >> 1
	all := tree.Iterator(Lte(cmp(i)), nil)
	i++
	for all.Next() {
		if i != all.Item() {
			b.Fatal(i, " != ", all.Item())
		}
		i++
	}
	b.StopTimer()
}

func BenchmarkIntIterBefore(b *testing.B) {
	b.Skipf("Long running and memory heavy")
	b.StopTimer()
	tree, cmp := newIntTree()
	defer tree.Release()
	for i := 0; i < b.N; i++ {
		tree.Insert(i)
	}
	b.StartTimer()
	i := b.N >> 1
	all := tree.Iterator(nil, Gte(cmp(i)))
	i = 0
	for all.Next() {
		if i != all.Item() {
			b.Fatal(i, " != ", all.Item())
		}
		i++
	}
	if i != b.N>>1 {
		b.Fatalf("Expected %d as the largest node, not %d", b.N>>1, i)
	}
	b.StopTimer()
}

func BenchmarkFetch(b *testing.B) {
	for _, sz := range []int{1 << 4, 1 << 8, 1 << 16, 1 << 24} {
		b.Run(fmt.Sprintf("btree size %d", sz), func(b *testing.B) {
			b.StopTimer()
			tree, _ := newIntTree()
			defer tree.Release()
			for i := 0; i < sz; i++ {
				tree.Insert(i)
			}
			fetched := 0
			items := rand.Perm(sz)
			b.StartTimer()
			for i := 0; i < b.N; i++ {
				if _, ok := tree.Fetch(items[i%sz] << 1); ok {
					fetched++
				}
			}
			b.StopTimer()
		})
		b.Run(fmt.Sprintf("map size %d", sz), func(b *testing.B) {
			b.StopTimer()
			m := map[int]struct{}{}
			for i := 0; i < sz; i++ {
				m[i] = struct{}{}
			}
			items := rand.Perm(sz)
			fetched := 0
			b.StartTimer()
			for i := 0; i < b.N; i++ {
				if _, ok := m[items[i%sz]<<1]; ok {
					fetched++
				}
			}

		})
	}
}

func BenchmarkIntIterRange(b *testing.B) {
	b.Skipf("Long running and memory heavy")
	b.StopTimer()
	tree, cmp := newIntTree()
	defer tree.Release()
	for i := 0; i < b.N; i++ {
		tree.Insert(i)
	}
	b.StartTimer()
	start := b.N >> 2
	end := (b.N >> 1) + start
	all := tree.Iterator(Lt(cmp(start)), Gte(cmp(end)))
	i := start
	for all.Next() {
		if i != all.Item() {
			b.Fatal(i, " != ", all.Item())
		}
		i++
	}
	if i != end {
		b.Fatalf("Expected %d as the largest node, not %d", end, i)
	}
	b.StopTimer()
}

func TestAscendAfter(t *testing.T) {
	tree, cmp := newIntTree()
	defer tree.Release()
	backing := []int{4, 6, 1, 3}
	for _, i := range backing {
		tree.Insert(i)
	}
	var ary, expected []int
	ary = nil
	// inclusive
	tree.After(Lte(cmp(-1)), func(idx int) bool {
		ary = append(ary, idx)
		return true
	})
	expected = []int{1, 3, 4, 6}
	if !reflect.DeepEqual(ary, expected) {
		t.Fatalf("expected %v but got %v", expected, ary)
	}
	ary = nil
	// inclusive
	tree.After(Lt(cmp(3)), func(idx int) bool {
		ary = append(ary, idx)
		return true
	})
	expected = []int{3, 4, 6}
	if !reflect.DeepEqual(ary, expected) {
		t.Fatalf("expected %v but got %v", expected, ary)
	}
	ary = nil
	// exclusive
	tree.After(Lte(cmp(3)), func(idx int) bool {
		ary = append(ary, idx)
		return true
	})
	expected = []int{4, 6}
	if !reflect.DeepEqual(ary, expected) {
		t.Fatalf("expected %v but got %v", expected, ary)
	}
	ary = nil
	tree.After(Lt(cmp(2)), func(idx int) bool {
		ary = append(ary, idx)
		return true
	})
	expected = []int{3, 4, 6}
	if !reflect.DeepEqual(ary, expected) {
		t.Fatalf("expected %v but got %v", expected, ary)
	}
}

func TestAscendBefore(t *testing.T) {
	tree, cmp := newIntTree()
	defer tree.Release()
	backing := []int{4, 6, 1, 3}
	for _, i := range backing {
		tree.Insert(i)
	}
	var ary []int
	tree.Before(Gt(cmp(10)), func(idx int) bool {
		ary = append(ary, idx)
		return true
	})
	expected := []int{1, 3, 4, 6}
	if !reflect.DeepEqual(ary, expected) {
		t.Fatalf("expected %v but got %v", expected, ary)
	}
	ary = nil
	tree.Before(Gte(cmp(4)), func(idx int) bool {
		ary = append(ary, idx)
		return true
	})
	expected = []int{1, 3}
	if !reflect.DeepEqual(ary, expected) {
		t.Fatalf("expected %v but got %v", expected, ary)
	}
	ary = nil
	tree.Before(Gt(cmp(4)), func(i int) bool {
		ary = append(ary, i)
		return true
	})
	expected = []int{1, 3, 4}
	if !reflect.DeepEqual(ary, expected) {
		t.Fatalf("expected %v but got %v", expected, ary)
	}
}
