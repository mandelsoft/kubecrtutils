package mapping

import (
	"maps"

	"github.com/mandelsoft/kubecrtutils/types/plain"
	"k8s.io/apimachinery/pkg/util/sets"
)

type Mappings map[string]string

func (m Mappings) Add(local, tgt string) {
	m[local] = tgt
}

func (m Mappings) Map(local string) string {
	if m == nil || m[local] == "" {
		return local
	}
	return m[local]
}

func (m Mappings) GetMapped(set plain.NameSet) plain.NameSet {
	if len(m) == 0 {
		return set
	}
	mapped := plain.NewNameSet()
	for n := range m {
		mapped.Add(m.Map(n))
	}
	return mapped
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
	for local, global := range add {
		if r[local] == "" {
			r[local] = global
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
	ComponentMappings() Mappings
	IndexMappings() Mappings

	IsNone() bool
	IsClustersNone() bool
	IsIndicesNone() bool
	IsComponentsNone() bool
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

func (n none) ComponentMappings() Mappings {
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

func (n none) IsComponentsNone() bool {
	return true
}

////////////////////////////////////////////////////////////////////////////////

func NewControllerMappings(clusters Mappings) *_cmappings {
	if clusters == nil {
		clusters = Mappings{}
	}
	return &_cmappings{
		indices:    Mappings{},
		clusters:   clusters,
		components: Mappings{},
	}
}

type _cmappings struct {
	indices    Mappings
	clusters   Mappings
	components Mappings
}

func (d *_cmappings) ClusterMappings() Mappings {
	return d.clusters
}
func (d *_cmappings) ComponentMappings() Mappings {
	return d.components
}
func (d *_cmappings) IndexMappings() Mappings {
	return d.indices
}

func (d *_cmappings) IsNone() bool {
	if d == nil {
		return true
	}
	return (d.indices == nil || len(d.indices) == 0) &&
		(d.clusters == nil || len(d.clusters) == 0) &&
		(d.components == nil || len(d.components) == 0)
}

func (d *_cmappings) IsComponentsNone() bool {
	if d == nil {
		return true
	}
	return d.components == nil || len(d.components) == 0
}

func (d *_cmappings) IsClustersNone() bool {
	if d == nil {
		return true
	}
	return d.clusters == nil || len(d.clusters) == 0
}

func (d *_cmappings) IsIndicesNone() bool {
	if d == nil {
		return true
	}
	return d.indices == nil || len(d.indices) == 0
}

func (m *_cmappings) ApplyTo(add ControllerMappings) ControllerMappings {
	if add == nil {
		return m
	}
	return &_cmappings{
		indices:    m.indices.ApplyTo(add.ClusterMappings()),
		components: m.components.ApplyTo(add.ComponentMappings()),
		clusters:   m.clusters.ApplyTo(add.ClusterMappings()),
	}
}
