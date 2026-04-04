package cacheindex

import (
	"context"
	"fmt"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/types"
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

// CreateIndices creates the indices for not disabled clusters.
func (d *_definitions) CreateIndices(ctx context.Context, mgr types.ControllerManager) error {
	if d.GetError() != nil {
		return d.GetError()
	}

	clusters := mgr.GetClusters()

	for n, i := range d.Elements {
		if clusters.IsDisabled(i.GetTarget()) {
			continue
		}
		err := i.Apply(ctx, nil, mgr)
		if err != nil {
			return fmt.Errorf("index %q: %w", n, err)
		}
	}
	return nil
}
