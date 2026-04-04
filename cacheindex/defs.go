package cacheindex

import (
	"context"
	"fmt"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/mandelsoft/logging"
)

func From(opts flagutils.OptionSetProvider) Definitions {
	return flagutils.GetFrom[Definitions](opts)
}

type Definitions = types.IndexDefinitions

type _definitions struct {
	internal.DefinitionsImpl[Definition, Definitions]
}

func NewDefinitions() Definitions {
	d := &_definitions{}
	d.DefinitionsImpl = internal.NewDefinitions[Definition, Definitions]("index", d)
	return d
}

// GetIndices creates the indices for not disabled clusters.
func (d *_definitions) GetIndices(ctx context.Context, clusters Clusters, logger logging.Logger) (Indices, error) {
	if d.GetError() != nil {
		return nil, d.GetError()
	}
	indices := NewIndices()
	for n, i := range d.Elements {
		if clusters.IsDisabled(i.GetTarget()) {
			continue
		}
		idx, err := i.Apply(ctx, clusters, logger)
		if err != nil {
			return nil, fmt.Errorf("index %q: %w", n, err)
		}
		indices.Add(idx)
	}
	return indices, nil
}

func (d *_definitions) ApplyMappings(mappings mapping.ControllerMappings) Definitions {
	defs := NewDefinitions()
	for _, e := range d.Elements {
		defs.Add(e.ApplyMappings(mappings))
	}
	return defs
}
