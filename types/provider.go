package types

import (
	"context"

	"github.com/mandelsoft/kubecrtutils/mapping"
)

type Applyable interface {
	Apply(ctx context.Context, mappings mapping.ControllerMappings, mgr ControllerManager) error
}

type IndexProvider interface {
	// CreateIndices creates and exports locally defined indices prior to creation of other elemens.
	CreateIndices(ctx context.Context, mapping mapping.ControllerMappings, mgr ControllerManager) error
}
