package ctrlmgmt

import (
	"context"

	"github.com/mandelsoft/kubecrtutils/types"
)

type _controllerMappings struct {
	indices  Mappings
	clusters Mappings
}

var _ types.ControllerMappings = (*_controllerMappings)(nil)

func (m *_controllerMappings) ClusterMappings() types.Mappings {
	return m.clusters
}

func (m *_controllerMappings) IndexMappings() types.Mappings {
	return m.indices
}

func (m *_controllerMappings) ApplyTo(add types.ControllerMappings) types.ControllerMappings {
	if add == nil {
		return m
	}
	return &_controllerMappings{
		indices:  m.indices.ApplyTo(add.ClusterMappings()),
		clusters: m.clusters.ApplyTo(add.ClusterMappings()),
	}
}

type _mappedController struct {
	ControllerDefinition
	_controllerMappings
}

func WithMappings(def types.ControllerDefinition) MappedControllerDefinition {
	return &_mappedController{
		ControllerDefinition: def,
		_controllerMappings: _controllerMappings{
			clusters: Mappings{},
			indices:  Mappings{},
		},
	}
}

// MapCluster maps a cluster name as used in the controller definition to
// a global controller manager cluster, when composing a controller
// set for a controller manager.
func (d *_mappedController) MapCluster(src, tgt string) MappedControllerDefinition {
	d.clusters[src] = tgt
	return d
}

func (d *_mappedController) CreateIndices(ctx context.Context, mapping types.ControllerMappings, mgr ControllerManager) error {
	return d.ControllerDefinition.CreateIndices(ctx, d.ApplyTo(mapping), mgr)
}

func (d *_mappedController) CreateController(ctx context.Context, mapping types.ControllerMappings, mgr ControllerManager) (types.Controller, error) {
	return d.ControllerDefinition.CreateController(ctx, d.ApplyTo(mapping), mgr)
}

func (d *_mappedController) MapIndex(src, tgt string) MappedControllerDefinition {
	d.indices[src] = tgt
	return d
}
