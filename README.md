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
    BenchmarkInsertIntSeq-10           	 8666037	       138.4 ns/op	      42 B/op	       0 allocs/op
    BenchmarkInsertIntSeqReverse-10    	11037992	       114.4 ns/op	      43 B/op	       0 allocs/op
    BenchmarkInsertIntRand-10          	 2686587	       756.9 ns/op	      30 B/op	       0 allocs/op
    BenchmarkDeleteIntSeq-10           	18283369	        74.21 ns/op	      29 B/op	       0 allocs/op
    BenchmarkDeleteIntRand-10          	 2703058	       796.6 ns/op	      24 B/op	       0 allocs/op
    BenchmarkInsertStringSeq-10        	 6777834	       182.9 ns/op	      48 B/op	       1 allocs/op
    BenchmarkInsertStringRand-10       	 1634024	       909.2 ns/op	      48 B/op	       1 allocs/op
    BenchmarkDeleteStringSeq-10        	13391804	        91.18 ns/op	      20 B/op	       0 allocs/op
    BenchmarkDeleteStringRand-10       	 1757889	       881.9 ns/op	      19 B/op	       0 allocs/op
    BenchmarkFetch/btree_size_16-10    	94638462	        12.93 ns/op
    BenchmarkFetch/map_size_16-10      	198753306	         6.014 ns/op
    BenchmarkFetch/btree_size_256-10   	46076608	        23.35 ns/op
    BenchmarkFetch/map_size_256-10     	152236431	         8.373 ns/op
    BenchmarkFetch/btree_size_65536-10 	13597803	        87.44 ns/op
    BenchmarkFetch/map_size_65536-10   	61784941	        19.08 ns/op
    BenchmarkFetch/btree_size_16777216-10         	 2100366	       582.8 ns/op
    BenchmarkFetch/map_size_16777216-10           	20900006	        57.07 ns/op
    PASS
    ok  	github.com/VictorLowther/btree	77.503s

Interestingly enough, the slowdown on the random benchmarks appears to be due to
branch misprediction rather than tree rebalancing performing more work -- dealing
with sorted data actually performs more rebalancing than random data.
