package mapping

import (
	"github.com/mandelsoft/goutils/set"

	. "github.com/mandelsoft/kubecrtutils/types/plain"
)

type Consumer interface {
	GetClusters() ClusterNames
	GetComponents() ComponentNames
	// GetRequiredClusters(mappings types.ControllerMappings) ClusterNames
}

type BaseMappable[S Consumer] interface {
	ControllerMappings

	GetRequiredClusters(mappings ControllerMappings) ClusterNames
	GetRequiredComponents(mappings ControllerMappings) ComponentNames
	ApplyTo(mappings ControllerMappings) ControllerMappings
}

type ClusterMappable[S Consumer] interface {
	MapCluster(src, tgt string) S
}
type IndexMappable[S Consumer] interface {
	MapIndex(src, tgt string) S
}
type ComponentMappable[S Consumer] interface {
	MapComponent(src, tgt string) S
}

type Mappable[S Consumer] interface {
	BaseMappable[S]
	ClusterMappable[S]
	IndexMappable[S]
	ComponentMappable[S]
}

type BaseMappings[S Consumer] struct {
	indices    Mappings
	clusters   Mappings
	components Mappings
	self       S
}

var (
	_ Mappable[Consumer]     = (*BaseMappings[Consumer])(nil)
	_ ControllerMappings     = (*BaseMappings[Consumer])(nil)
	_ BaseMappable[Consumer] = (*BaseMappings[Consumer])(nil)
)

func NewBaseMappings[S Consumer](self S) *BaseMappings[S] {
	return &BaseMappings[S]{
		self:       self,
		indices:    Mappings{},
		clusters:   Mappings{},
		components: Mappings{},
	}
}

func (d *BaseMappings[S]) GetSelf() S {
	return d.self
}

func (d *BaseMappings[S]) IsNone() bool {
	if d == nil {
		return true
	}
	return (d.indices == nil || len(d.indices) == 0) &&
		(d.clusters == nil || len(d.clusters) == 0) &&
		(d.components == nil || len(d.components) == 0)
}

func (d *BaseMappings[S]) IsComponentsNone() bool {
	if d == nil {
		return true
	}
	return d.components == nil || len(d.components) == 0
}

func (d *BaseMappings[S]) IsClustersNone() bool {
	if d == nil {
		return true
	}
	return d.clusters == nil || len(d.clusters) == 0
}

func (d *BaseMappings[S]) IsIndicesNone() bool {
	if d == nil {
		return true
	}
	return d.indices == nil || len(d.indices) == 0
}

func (d *BaseMappings[S]) MapIndex(src, tgt string) S {
	d.indices[src] = tgt
	return d.self
}

func (d *BaseMappings[S]) MapComponent(src, tgt string) S {
	d.components[src] = tgt
	return d.self
}

// MapCluster maps a cluster name as used in the controller definition to
// a global controller manager cluster, when composing a controller
// set for a controller manager.
func (d *BaseMappings[S]) MapCluster(src, tgt string) S {
	d.clusters[src] = tgt
	return d.self
}

func (m *BaseMappings[S]) ComponentMappings() Mappings {
	return m.components
}

func (m *BaseMappings[S]) ClusterMappings() Mappings {
	return m.clusters
}

func (m *BaseMappings[S]) IndexMappings() Mappings {
	return m.indices
}

func (m *BaseMappings[S]) ApplyTo(add ControllerMappings) ControllerMappings {
	if add == nil {
		return m
	}
	return &BaseMappings[S]{
		indices:    m.indices.ApplyTo(add.ClusterMappings()),
		components: m.components.ApplyTo(add.ComponentMappings()),
		clusters:   m.clusters.ApplyTo(add.ClusterMappings()),
	}
}

func (m *BaseMappings[S]) GetRequiredClusters(mappings ControllerMappings) ClusterNames {
	names := set.New[string]()
	mp := m.ApplyTo(mappings).ClusterMappings()
	for n := range m.self.GetClusters() {
		names.Add(mp.Map(n))
	}
	return names
}

func (m *BaseMappings[S]) GetRequiredComponents(mappings ControllerMappings) ComponentNames {
	names := set.New[string]()
	mp := m.ApplyTo(mappings).ComponentMappings()
	for n := range m.self.GetComponents() {
		names.Add(mp.Map(n))
	}
	return names
}
