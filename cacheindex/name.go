package cacheindex

import (
	"strings"

	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/types"
)

func ComposeName(name, cluster string) string {
	return name + ":" + cluster
}

func SplitName(name string) (string, string) {
	fields := strings.Split(name, ":")
	if len(fields) == 1 {
		return fields[0], ""
	}
	return fields[0], fields[1]
}

func BaseName(name string) string {
	n, _ := SplitName(name)
	return n
}

func ClusterName(name string) string {
	_, c := SplitName(name)
	return c
}

func ComposeNameFor(def Definition, mappings mapping.ControllerMappings, mgr types.ControllerManager) string {
	mappings = mapping.DefaultMappings(mappings)
	return ComposeName(mappings.IndexMappings().Map(def.GetName()), mgr.GetClusters().Get(mappings.ClusterMappings().Map(def.GetTarget())).GetEffective().GetName())
}

func MapName(name string, mappings mapping.ControllerMappings) string {
	n, c := SplitName(name)
	if c == "" {
		return mappings.IndexMappings().Map(name)
	}
	return ComposeName(mappings.IndexMappings().Map(n), mappings.ClusterMappings().Map(c))
}
