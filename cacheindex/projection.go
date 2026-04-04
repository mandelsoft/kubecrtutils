package cacheindex

import (
	"fmt"

	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils/mapping"
)

type IndexNames = set.Set[string]

func Map(indices Indices, mapping mapping.Mappings, names IndexNames) (Indices, error) {
	n := NewIndices()

	for local := range names {
		global := mapping.Map(local)
		c := indices.Get(global)
		if c == nil {
			return nil, fmt.Errorf("global index %q for %q not defined", global, local)
		}
		n.Add(NewAlias(local, c))
	}
	return n, nil
}
