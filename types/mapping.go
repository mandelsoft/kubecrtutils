package types

import (
	"maps"

	"k8s.io/apimachinery/pkg/util/sets"
)

type Mappings map[string]string

func (m Mappings) Map(local string) string {
	if m == nil || m[local] == "" {
		return local
	}
	return m[local]
}

func (m Mappings) ApplyTo(add Mappings) Mappings {
	if m == nil {
		return add
	}
	if add == nil {
		return m
	}
	r := maps.Clone(add)
	for local, global := range m {
		if r[local] == "" {
			r[local] = add.Map(global)
		}
	}
	return r
}

func IdentityMapping(set sets.Set[string]) Mappings {
	m := Mappings{}
	for n := range set {
		m[n] = n
	}
	return m
}

type ControllerMappings interface {
	ClusterMappings() Mappings
	IndexMappings() Mappings

	IsNone() bool
	IsClustersNone() bool
	IsIndicesNone() bool
}

func DefaultMappings(mappings ControllerMappings) ControllerMappings {
	if mappings == nil {
		return NoMappings()
	}
	return mappings
}

func NoMappings() ControllerMappings {
	return none{}
}

type none struct{}

func (n none) ClusterMappings() Mappings {
	return nil
}

func (n none) IndexMappings() Mappings {
	return nil
}

func (n none) IsNone() bool {
	return true
}

func (n none) IsClustersNone() bool {
	return true
}

func (n none) IsIndicesNone() bool {
	return true
}
