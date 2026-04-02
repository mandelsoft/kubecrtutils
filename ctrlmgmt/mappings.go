package ctrlmgmt

import (
	"context"

	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/controller"
	"github.com/mandelsoft/kubecrtutils/types"
)

type clusterconsumer interface {
	GetClusters() ClusterNames
	// GetRequiredClusters(mappings types.ControllerMappings) ClusterNames
}

type mappable[S any] interface {
	MapCluster(src, tgt string) S
	MapIndex(src, tgt string) S
}

type _mappings[S clusterconsumer] struct {
	indices  Mappings
	clusters Mappings
	self     S
}

var (
	_ mappable[controller.Definition] = (*_mappings[controller.Definition])(nil)
	_ types.ControllerMappings        = (*_mappings[controller.Definition])(nil)
)

func (d *_mappings[S]) IsNone() bool {
	if d == nil {
		return true
	}
	return (d.indices == nil || len(d.indices) == 0) && (d.clusters == nil || len(d.clusters) == 0)
}

func (d *_mappings[S]) IsClustersNone() bool {
	if d == nil {
		return true
	}
	return d.clusters == nil || len(d.clusters) == 0
}

func (d *_mappings[S]) IsIndicesNone() bool {
	if d == nil {
		return true
	}
	return d.indices == nil || len(d.indices) == 0
}

func (d *_mappings[S]) MapIndex(src, tgt string) S {
	d.indices[src] = tgt
	return d.self
}

// MapCluster maps a cluster name as used in the controller definition to
// a global controller manager cluster, when composing a controller
// set for a controller manager.
func (d *_mappings[S]) MapCluster(src, tgt string) S {
	d.clusters[src] = tgt
	return d.self
}

func (m *_mappings[S]) ClusterMappings() types.Mappings {
	return m.clusters
}

func (m *_mappings[S]) IndexMappings() types.Mappings {
	return m.indices
}

func (m *_mappings[S]) ApplyTo(add types.ControllerMappings) types.ControllerMappings {
	if add == nil {
		return m
	}
	return &_mappings[S]{
		indices:  m.indices.ApplyTo(add.ClusterMappings()),
		clusters: m.clusters.ApplyTo(add.ClusterMappings()),
	}
}

func (m *_mappings[S]) GetRequiredClusters(mappings types.ControllerMappings) ClusterNames {
	names := set.New[string]()
	mp := m.ApplyTo(mappings).ClusterMappings()
	for n := range m.self.GetClusters() {
		names.Add(mp.Map(n))
	}
	return names
}

////////////////////////////////////////////////////////////////////////////////

func WithMappings(def controller.Definition) MappedControllerDefinition {
	m := &_mappedController{
		Definition: def,
		_mappings: _mappings[*_mappedController]{
			clusters: Mappings{},
			indices:  Mappings{},
		},
	}
	m._mappings.self = m
	return m
}

////////////////////////////////////////////////////////////////////////////////

type _mappedController struct {
	controller.Definition
	_mappings[*_mappedController]
}

var _ controller.Definition = (*_mappedController)(nil)

func (d *_mappedController) GetRequiredClusters(mappings types.ControllerMappings) ClusterNames {
	// resolve method
	return d._mappings.GetRequiredClusters(mappings)
}

func (d *_mappedController) GetForeignIndices() cacheindex.Definitions {
	return d.Definition.GetForeignIndices().ApplyMappings(d)
}

func (d *_mappedController) CreateIndices(ctx context.Context, mapping types.ControllerMappings, mgr ControllerManager) error {
	return d.Definition.CreateIndices(ctx, d.ApplyTo(mapping), mgr)
}

func (d *_mappedController) CreateController(ctx context.Context, mapping types.ControllerMappings, mgr ControllerManager) (types.Controller, error) {
	return d.Definition.CreateController(ctx, d.ApplyTo(mapping), mgr)
}
