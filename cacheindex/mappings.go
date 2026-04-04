package cacheindex

import (
	"context"

	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/types"
)

// use simplified mapping methods
// replaced dtype _mapped

// should be consistent with controller/mappings.go

////////////////////////////////////////////////////////////////////////////////

func (d *_mapped) GetRequiredClusters(mappings mapping.ControllerMappings) types.ClusterNames {
	// resolve method
	return d.Mapped.GetRequiredClusters(mappings)
}

func (d *_mapped) GetRequiredComponents(mappings mapping.ControllerMappings) types.ComponentNames {
	// resolve method
	return d.Mapped.GetRequiredComponents(mappings)
}

func (d *_mapped) Apply(ctx context.Context, mappings mapping.ControllerMappings, mgr types.ControllerManager) error {
	return d.Definition.Apply(ctx, d.ApplyTo(mappings), mgr)
}
