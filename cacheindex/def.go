package cacheindex

import (
	"context"
	"fmt"

	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/mandelsoft/logging"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IndexerFunc[T client.Object] = func(T) []string

type Definition interface {
	GetName() string
	GetTarget() string
	GetIndexerFunc() client.IndexerFunc
	Apply(ctx context.Context, set Clusters, logger logging.Logger) (types.Index, error)
}

type TypedDefinition[T any, P kubecrtutils.ObjectPointer[T]] interface {
	Definition
	TypedApply(ctx context.Context, set types.Clusters, logger logging.Logger) (TypedIndex[T], error)
}

type _definition[T any, P kubecrtutils.ObjectPointer[T]] struct {
	internal.Element
	target  string
	proto   client.Object
	idxfunc IndexerFunc[P]
}

func NewDefinition[T any, P kubecrtutils.ObjectPointer[T]](name string, target string, idxfunc IndexerFunc[P]) TypedDefinition[T, P] {
	return &_definition[T, P]{
		Element: internal.NewElement(name),
		target:  target,
		idxfunc: idxfunc,
		proto:   kubecrtutils.Proto[T, P](),
	}
}

func (d *_definition[T, P]) GetTarget() string {
	return d.target
}

func (d *_definition[T, P]) GetIndexerFunc() client.IndexerFunc {
	return d.indexer
}

func (d *_definition[T, P]) indexer(obj client.Object) []string {
	return d.idxfunc(obj.(any).(P))

}

func (d *_definition[T, P]) Apply(ctx context.Context, set Clusters, logger logging.Logger) (types.Index, error) {
	c := set.Get(d.GetTarget())
	if c == nil {
		return nil, fmt.Errorf("cluster %q not found", d.GetTarget())
	}

	gk, err := kubecrtutils.GKForObject(c, d.proto)
	if err != nil {
		return nil, err
	}

	logger.Info("creating index {{index}} for {{resource}} on {{cluster}}[{{effcluster}}]", "index", d.GetName(), "resource", gk, "cluster", d.GetTarget(), "effcluster", c.GetEffective().GetName())
	idx, err := c.CreateIndex(ctx, d.GetName(), d.proto, d.indexer, func(_c ClusterEquivalent, name string) (Index, error) {
		idx, err := NewDefaultIndex(d.GetName(), _c, d.proto)
		if err != nil {
			return nil, err
		}
		return &_typedIndex[T]{
			idx,
		}, nil
	})

	if err != nil {
		return nil, err
	}
	return idx, nil
}

func (d *_definition[T, P]) TypedApply(ctx context.Context, set Clusters, logger logging.Logger) (TypedIndex[T], error) {
	idx, err := d.Apply(ctx, set, logger)
	if err != nil {
		return nil, err
	}
	return idx.(TypedIndex[T]), nil
}
