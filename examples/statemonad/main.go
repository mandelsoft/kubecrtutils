package main

import (
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////

// First returns the first result of a sequence of multiple function results.
func First[T any](v T, rest ...any) T {
	return v
}

// Second returns the second result of a sequence of multiple function results..
func Second[T any](a any, v T, rest ...any) T {
	return v
}

////////////////////////////////////////////////////////////////////////////////

type BinTree[V any] struct {
	left  *BinTree[V]
	value V
	right *BinTree[V]
}

func (b *BinTree[V]) String() string {
	if b.IsLeaf() {
		return "-"
	}

	return fmt.Sprintf("(%s/%v/%s)", b.left, b.value, b.right)
}

func (b *BinTree[V]) IsLeaf() bool {
	return b.left == nil && b.right == nil
}

type Numbered[V any] struct {
	index int
	value V
}

func (n Numbered[V]) String() string {
	return fmt.Sprintf("[%d: %v]", n.index, n.value)
}

func Leaf[V any]() *BinTree[V] {
	return &BinTree[V]{}
}

func Branch[V any](left *BinTree[V], v V, right *BinTree[V]) *BinTree[V] {
	return &BinTree[V]{
		left:  left,
		value: v,
		right: right,
	}
}

////////////////////////////////////////////////////////////////////////////////

type Unit *struct{}

type State[S, V any] func(S) (S, V)

// Monad functions for type State
// ok = pure
// addThen = bind

func ok[S, V any](v V) State[S, V] {
	return func(s S) (S, V) {
		return s, v
	}
}

func andThen[S, A, B any](first State[S, A], next func(A) State[S, B]) State[S, B] {
	return func(s S) (S, B) {
		st, a := first(s)
		return next(a)(st)
	}
}

/////////////////////////////////
// helper functions to set and get state

func get[S any](s S) (S, S) {
	return s, s
}

func set[S any](s S) State[S, Unit] {
	return func(_ S) (S, Unit) {
		return s, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// Now the operations implemented using State monad.

func NumberedTree[V any](t *BinTree[V]) *BinTree[Numbered[V]] {
	return Second(helper(t)(0))
}

func helper[V any](t *BinTree[V]) State[int, *BinTree[Numbered[V]]] {
	if t.IsLeaf() {
		return ok[int](Leaf[Numbered[V]]())
	} else {
		// type parameters not required, but shown for clarification
		return andThen[int, *BinTree[Numbered[V]], *BinTree[Numbered[V]]](
			helper(t.left),
			func(nleft *BinTree[Numbered[V]]) State[int, *BinTree[Numbered[V]]] {
				return andThen[int, int, *BinTree[Numbered[V]]](
					get[int],
					func(n int) State[int, *BinTree[Numbered[V]]] {
						return andThen[int, Unit, *BinTree[Numbered[V]]](
							set(n+1),
							func(Unit) State[int, *BinTree[Numbered[V]]] {
								return andThen[int, *BinTree[Numbered[V]], *BinTree[Numbered[V]]](
									helper(t.right),
									func(nright *BinTree[Numbered[V]]) State[int, *BinTree[Numbered[V]]] {
										mt := Branch[Numbered[V]](nleft, Numbered[V]{n, t.value}, nright)
										return ok[int](mt)
									},
								)
							},
						)
					},
				)
			},
		)
	}
}

func main() {

	tree := Branch(Branch(Leaf[int](), 3, Leaf[int]()), 5, Branch(Leaf[int](), 7, Leaf[int]()))

	fmt.Printf("%+v\n", tree)
	fmt.Printf("%+v\n", NumberedTree(tree))
}

// def number (t : BinTree α) : BinTree (Nat × α) :=
//  let rec helper : BinTree α → State Nat (BinTree (Nat × α))
//    | BinTree.leaf => pure BinTree.leaf
//    | BinTree.branch left x right => do
//      let numberedLeft ← helper left
//      let n ← get
//      set (n + 1)
//      let numberedRight ← helper right
//      ok (BinTree.branch numberedLeft (n, x) numberedRight)
//  (helper t 0).snd
