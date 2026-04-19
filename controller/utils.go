package controller

import (
	"slices"

	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils/types"
)

type ControllerRefrences map[string]set.Set[string]

func (a ControllerRefrences) AddAll(n string, sub set.Set[string]) {
	s, ok := a[n]
	if !ok {
		s = set.New[string]()
		a[n] = s
	}
	s.AddAll(sub)
}

func (a ControllerRefrences) Add(n, c string) {
	s, ok := a[n]
	if !ok {
		s = set.New[string](c)
		a[n] = s
	} else {
		s.Add(c)
	}
}

func CompleteSet(cs types.ControllerSet) map[string]set.Set[string] {
	r := ControllerRefrences{}

	for _, n := range cs.GetNames() {
		r.Add(n, n)
	}
	names := cs.GetNames()
	groups := cs.GetGroups()
	for g, list := range groups {
		handle(r, names, groups, g, list)
	}
	return r
}

func handle(r ControllerRefrences, names []string, groups map[string][]string, name string, list []string) {
	for _, n := range list {
		if r[n] == nil {
			if sub, ok := groups[n]; ok {
				handle(r, names, groups, n, sub)
				r.AddAll(name, r[n])
			}
		}
		if slices.Contains(names, n) {
			r.Add(name, n)
		}
	}
}
