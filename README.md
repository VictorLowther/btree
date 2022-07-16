# btree

btree is an implementation of generic balanced binary trees in Go. 
This packages provides generic AVL trees. 

## Why

AVL trees provide stricter balance guarantees than red-black trees, making search and iteration over
a subset of the tree on average faster.  Additionally, the balance algorithms handle inserting
and removing data in a semi-sorted fashion better than red-black trees do, at the slight cost of worse performance
when inserting and removing random data.

## Installation

go get https://github.com/VictorLowther/btree

## Example

    package main
    import github.com/VictorLowther/btree
    import fmt

    func main() {
        tree := btree.New[int](func(a,b int) {return a < b})
        for i := 0; i < 10; i++ {
            tree.Insert(i)
        }
        tree.Reverse()
        iter := tree.Iterate(nil, nil)
        for iter.Next() {
            fmt.Println(iter.Item())
        }
    }

## Benchmarks:

On a Macbook Pro M1 Max:

    % go test -bench .
    goos: darwin
    goarch: arm64
    pkg: github.com/VictorLowther/btree
    BenchmarkInsertIntSeq-10               8886606       139.6 ns/op      42 B/op       0 allocs/op
    BenchmarkInsertIntSeqReverse-10       11320531       113.6 ns/op      43 B/op       0 allocs/op
    BenchmarkInsertIntRand-10              2643009       826.5 ns/op      29 B/op       0 allocs/op
    BenchmarkDeleteIntSeq-10              18383584       73.93 ns/op      29 B/op       0 allocs/op
    BenchmarkDeleteIntRand-10              2096901       748.9 ns/op      16 B/op       0 allocs/op
    BenchmarkInsertStringSeq-10            6041194       198.3 ns/op      48 B/op       1 allocs/op
    BenchmarkInsertStringRand-10           1618108       994.1 ns/op      48 B/op       1 allocs/op
    BenchmarkDeleteStringSeq-10           14941672       92.09 ns/op      17 B/op       0 allocs/op
    BenchmarkDeleteStringRand-10           1652192       858.1 ns/op      20 B/op       0 allocs/op
    BenchmarkIntIterAll-10               231067432       5.153 ns/op       0 B/op       0 allocs/op
    BenchmarkFetch/btree_size_16-10       97148944       12.96 ns/op       0 B/op       0 allocs/op
    BenchmarkFetch/map_size_16-10        200721495       6.077 ns/op
    BenchmarkFetch/btree_size_256-10      40678252       27.01 ns/op       0 B/op       0 allocs/op
    BenchmarkFetch/map_size_256-10       152659923       7.475 ns/op
    BenchmarkFetch/btree_size_65536-10    13663353       86.88 ns/op       0 B/op       0 allocs/op
    BenchmarkFetch/map_size_65536-10      61720060       21.16 ns/op
    BenchmarkFetch/btree_size_16777216-10  2030596       596.1 ns/op       0 B/op       0 allocs/op
    BenchmarkFetch/map_size_16777216-10   20909883       57.17 ns/op
    PASS
    ok  	github.com/VictorLowther/btree	79.935s

Interestingly enough, the slowdown on the random benchmarks appears to be due to
branch misprediction rather than tree rebalancing performing more work -- dealing
with sorted data actually performs more rebalancing than random data.
