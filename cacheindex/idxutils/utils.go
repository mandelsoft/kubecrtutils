package idxutils

import (
	"fmt"
	"reflect"

	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/mandelsoft/logging"
)

type Mapper = func(string) string

func ImportIndex[I cacheindex.Index](logger logging.Logger, i cacheindex.Definition, clusters types.Clusters, mapping mapping.ControllerMappings, mgr types.ControllerManager, local map[string]I, mapper Mapper) error {
	n := i.GetName()
	c := clusters.Get(i.GetTarget()).GetEffective()

	glob := cacheindex.MapName(n, mapping)

	// import indexer
	idx := mgr.GetIndex(glob)
	if idx == nil {
		return fmt.Errorf("imported index %q->%q not found", n, glob)
	}

	f := i.GetIndexer()
	if f == nil {
		logger.Info("  importing index {{index}}->{{global}}", "index", n, "global", glob)
	} else {
		logger.Info("  using local index {{index}}->{{global}}", "index", n, "global", glob)
	}
	if reflect.TypeOf(i.GetResource()) != reflect.TypeOf(idx.GetResource()) {
		return fmt.Errorf("index %q->%g resource type mismatch: expected %T, but found %T", n, glob, i.GetResource(), idx.GetResource())
	}
	if c != idx.GetCluster().GetEffective() {
		return fmt.Errorf("index %q->%q cluster mismatch: expected %s[%s], but found %s", n, glob, i.GetTarget(), c.GetEffective().GetName(), idx.GetCluster().GetEffective().GetName())
	}

	eff := generics.Cast[I](idx.GetEffective())
	if mapper != nil {
		n = mapper(n)
	}
	if _, ok := local[n]; ok {
		return fmt.Errorf("index %q->%q cluster mismatch: expected %s[%s], but found %s", n, glob, i.GetTarget(), c.GetEffective().GetName(), idx.GetCluster().GetEffective().GetName())

	}
	local[n] = eff
	return nil
}

func AddShortNames(local string, indices map[string]cacheindex.Index) {
	simple := map[string]int{}
	for n := range indices {
		b := cacheindex.BaseName(n)
		simple[b] = simple[b] + 1
	}

	for n := range indices {
		b, c := cacheindex.SplitName(n)
		if simple[b] == 1 || c == "" || c == local {
			indices[b] = indices[n]
		}
	}
}

func ImportIndices(all map[string]cacheindex.Index, logger logging.Logger, local string, clusters types.Clusters, mapping mapping.ControllerMappings, mgr types.ControllerManager, defs ...cacheindex.Definitions) error {
	for _, set := range defs {
		for _, i := range set.Elements {
			err := ImportIndex(logger, i, clusters, mapping, mgr, all, nil)
			if err != nil {
				return err
			}
		}
	}
	AddShortNames(local, all)

	if len(all) > 0 {
		logger.Info("  available index names:")
		for n, i := range all {
			if cacheindex.ClusterName(n) == "" {
				for c := range all {
					b, cl := cacheindex.SplitName(c)
					if cl != "" && n == b {
						logger.Info("  - {{importedindex}} -> {{mapped}} -> {{effective}}", "importedindex", n, "mapped", c, "effective", i.GetEffective().GetName())
						break
					}
				}
			} else {
				logger.Info("  - {{importedindex}} -> {{effective}}", "importedindex", n, "effective", i.GetEffective().GetName())
			}
		}
	} else {
		logger.Info("  no indices used")
	}

	return nil
}
