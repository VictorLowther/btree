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

    % go test -v -bench .  |&tee test.log
    === RUN   TestRotate
    btree_test.go:130: Tree populated correctly
    btree_test.go:150: Tree rotated left correctly
    btree_test.go:171: Tree rotated right correctly
    --- PASS: TestRotate (0.00s)
    === RUN   TestCases
    --- PASS: TestCases (0.00s)
    === RUN   TestRange
    --- PASS: TestRange (0.00s)
    === RUN   TestIter
    --- PASS: TestIter (0.00s)
    === RUN   TestRandomInsertOrder
    --- PASS: TestRandomInsertOrder (0.36s)
    === RUN   TestRandomInsertDelete
    btree_test.go:311: ins: 10000, rebalances/ins: 0.472000, rms: 10000, rebalances/rm: 0.277800
    --- PASS: TestRandomInsertDelete (0.66s)
    === RUN   TestRandomInsertDeleteNonExistent
    btree_test.go:345: ins: 100, rebalances/ins: 0.470000, rms: 100, rebalances/rm: 0.520000
    --- PASS: TestRandomInsertDeleteNonExistent (0.00s)
    === RUN   TestRandomInsertPartialDeleteOrder
    btree_test.go:367: ins: 1000, rebalances/ins: 0.454000, rms: 1000, rebalances/rm: 0.562000
    --- PASS: TestRandomInsertPartialDeleteOrder (0.01s)
    === RUN   TestRandomInsertStats
    btree_test.go:383: ins: 100000, rebalances/ins: 0.465250, rms: 0, rebalances/rm: NaN
    --- PASS: TestRandomInsertStats (0.03s)
    === RUN   TestSeqInsertStats
    btree_test.go:399: ins: 100000, rebalances/ins: 0.999830, rms: 0, rebalances/rm: NaN
    --- PASS: TestSeqInsertStats (0.01s)
    === RUN   TestAscendAfter
    --- PASS: TestAscendAfter (0.00s)
    === RUN   TestAscendBefore
    --- PASS: TestAscendBefore (0.00s)
    goos: darwin
    goarch: arm64
    pkg: github.com/VictorLowther/btree
    BenchmarkInsertIntSeq
    BenchmarkInsertIntSeq-10           	 8760205	       140.7 ns/op
    BenchmarkInsertIntSeqReverse
    BenchmarkInsertIntSeqReverse-10    	11518032	       113.0 ns/op
    BenchmarkInsertIntRand
    BenchmarkInsertIntRand-10          	 2582887	       782.8 ns/op
    BenchmarkDeleteIntSeq
    BenchmarkDeleteIntSeq-10           	18287211	        73.69 ns/op
    BenchmarkDeleteIntRand
    BenchmarkDeleteIntRand-10          	 2558193	       901.9 ns/op
    BenchmarkInsertStringSeq
    BenchmarkInsertStringSeq-10        	 6735705	       194.2 ns/op
    BenchmarkInsertStringRand
    BenchmarkInsertStringRand-10       	 1000000	      1187 ns/op
    BenchmarkDeleteStringSeq
    BenchmarkDeleteStringSeq-10        	12886626	        90.69 ns/op
    BenchmarkDeleteStringRand
    BenchmarkDeleteStringRand-10       	 1434376	       933.4 ns/op
    PASS
    ok  	github.com/VictorLowther/btree	37.723s

Interestingly enough, the slowdown on the random benchmarks appears to be due to
branch misprediction rather than tree rebalancing performing more work -- dealing
with sorted data actually performs more rebalancing than random data.
