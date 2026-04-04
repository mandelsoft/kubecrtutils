package cacheindex

import (
	"strings"

	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/types"
)

func ComposeName(name, cluster string) string {
	return name + ":" + cluster
}

func ComposeNameFor(def Definition, mappings mapping.ControllerMappings, mgr types.ControllerManager) string {
	mappings = mapping.DefaultMappings(mappings)
	return ComposeName(mappings.IndexMappings().Map(def.GetName()), mgr.GetClusters().Get(mappings.ClusterMappings().Map(def.GetTarget())).GetEffective().GetName())
}

func MapName(name string, mappings mapping.ControllerMappings) string {
	fields := strings.Split(name, ":")
	if len(fields) == 1 {
		return mappings.IndexMappings().Map(name)
	}
	return ComposeName(mappings.IndexMappings().Map(fields[0]), mappings.ClusterMappings().Map(fields[1]))
}
