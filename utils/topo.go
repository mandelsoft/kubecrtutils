package utils

import (
	"iter"
	"slices"
)

type Dependencies[E any, I comparable] func(e E) iter.Seq2[I, E]

func TopoSort[E any, I comparable](seq2 iter.Seq2[I, E], d Dependencies[E, I]) (order []I, cycle []I) {
	var o []I

	for k, v := range seq2 {
		cycle := _order(k, v, d, nil, &o)
		if cycle != nil {
			return nil, cycle
		}
	}
	return o, nil
}

func _order[E any, I comparable](i I, e E, d Dependencies[E, I], history []I, o *[]I) []I {
	if idx := slices.Index(history, i); idx >= 0 {
		return append(history[idx:], i)
	}
	if !slices.Contains(*o, i) {
		history = append(history, i)
		for k, v := range d(e) {
			cycle := _order[E, I](k, v, d, history, o)
			if cycle != nil {
				return cycle
			}
		}
		*o = append(*o, i)
	}
	return nil
}
