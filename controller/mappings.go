package controller

import (
	"context"

	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/types"
)

// should be identical to component/mappings.go

func WithMappings(def Definition) *_mapped {
	m := &_mapped{
		Definition: def,
	}
	m.Mappable = mapping.NewBaseMappings(m)
	return m
}

////////////////////////////////////////////////////////////////////////////////

type _mapped struct {
	Definition
	mapping.Mappable[*_mapped]
}

var _ Definition = (*_mapped)(nil)

func (d *_mapped) GetRequiredClusters(mappings mapping.ControllerMappings) types.ClusterNames {
	// resolve method
	return d.Mappable.GetRequiredClusters(mappings)
}

func (d *_mapped) GetRequiredComponents(mappings mapping.ControllerMappings) types.ComponentNames {
	// resolve method
	return d.Mappable.GetRequiredComponents(mappings)
}

func (d *_mapped) CreateIndices(ctx context.Context, mappings mapping.ControllerMappings, mgr types.ControllerManager) error {
	return d.Definition.CreateIndices(ctx, d.ApplyTo(mappings), mgr)
}

func (d *_mapped) Apply(ctx context.Context, mappings mapping.ControllerMappings, mgr types.ControllerManager) error {
	return d.Definition.Apply(ctx, d.ApplyTo(mappings), mgr)
}
