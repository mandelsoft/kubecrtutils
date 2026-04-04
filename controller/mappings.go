package controller

import (
	"context"

	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/types"
)

// should be identical to component/mappings.go

func WithMappings(def Definition) *_mapped {
	m := &_mapped{
		Definition: def,
	}
	m.BaseMappings = mapping.NewBaseMappings(m)
	return m
}

////////////////////////////////////////////////////////////////////////////////

type _mapped struct {
	Definition
	*mapping.BaseMappings[*_mapped]
}

var _ Definition = (*_mapped)(nil)

func (d *_mapped) GetRequiredClusters(mappings mapping.ControllerMappings) types.ClusterNames {
	// resolve method
	return d.BaseMappings.GetRequiredClusters(mappings)
}

func (d *_mapped) GetRequiredComponents(mappings mapping.ControllerMappings) types.ComponentNames {
	// resolve method
	return d.BaseMappings.GetRequiredComponents(mappings)
}

func (d *_mapped) GetForeignIndices() cacheindex.Definitions {
	return d.Definition.GetForeignIndices().ApplyMappings(d)
}

func (d *_mapped) CreateIndices(ctx context.Context, mapping mapping.ControllerMappings, mgr types.ControllerManager) error {
	return d.Definition.CreateIndices(ctx, d.ApplyTo(mapping), mgr)
}

func (d *_mapped) Apply(ctx context.Context, mapping mapping.ControllerMappings, mgr types.ControllerManager) (Controller, error) {
	return d.Definition.Apply(ctx, d.ApplyTo(mapping), mgr)
}
