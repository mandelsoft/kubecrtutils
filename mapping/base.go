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

type ConfigurableMappings[S any] interface {
	ControllerMappings

	UseMappings(mappings ControllerMappings) S

	ClusterMappable[S]
	IndexMappable[S]
	ComponentMappable[S]
}

type BaseMappable[S Consumer] interface {
	ControllerMappings

	UseMappings(mappings ControllerMappings) S

	GetRequiredClusters(mappings ControllerMappings) ClusterNames
	GetRequiredComponents(mappings ControllerMappings) ComponentNames
	ApplyTo(mappings ControllerMappings) ControllerMappings
}

type ClusterMappable[S any] interface {
	MapCluster(src, tgt string) S
}
type IndexMappable[S any] interface {
	MapIndex(src, tgt string) S
}
type ComponentMappable[S any] interface {
	MapComponent(src, tgt string) S
}

type Mappable[S Consumer] interface {
	BaseMappable[S]

	ClusterMappable[S]
	IndexMappable[S]
	ComponentMappable[S]
}

type BaseMappings[S Consumer] struct {
	_mcmappings[S]
}

var (
	_ Mappable[Consumer]     = (*BaseMappings[Consumer])(nil)
	_ ControllerMappings     = (*BaseMappings[Consumer])(nil)
	_ BaseMappable[Consumer] = (*BaseMappings[Consumer])(nil)
)

func NewBaseMappings[S Consumer](self S) *BaseMappings[S] {
	return &BaseMappings[S]{
		_mcmappings: *newConfigurableMappings(self),
	}
}

func (d *BaseMappings[S]) GetSelf() S {
	return d.self
}

func (d *BaseMappings[S]) UseMappings(mappings ControllerMappings) S {
	for s, t := range mappings.IndexMappings() {
		d.indices[s] = t
	}
	for s, t := range mappings.ClusterMappings() {
		d.clusters[s] = t
	}
	for s, t := range mappings.ComponentMappings() {
		d.components[s] = t
	}
	return d.self
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
