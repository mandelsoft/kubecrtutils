package cacheindex

import (
	"context"
	"fmt"

	"github.com/mandelsoft/goutils/generics"
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
	GetResource() client.Object
	GetIndexerFunc() client.IndexerFunc
	Apply(ctx context.Context, set Clusters, logger logging.Logger) (types.Index, error)
}

type Reference interface {
	Definition
	IsRef() bool
}

type TypedDefinition[P kubecrtutils.ObjectPointer[T], T any] interface {
	Definition
	TypedApply(ctx context.Context, set types.Clusters, logger logging.Logger) (TypedIndex[T], error)
}

type _definition[P kubecrtutils.ObjectPointer[T], T any] struct {
	internal.Element
	target  string
	proto   client.Object
	idxfunc IndexerFunc[P]
}

type _reference[P kubecrtutils.ObjectPointer[T], T any] struct {
	_definition[P, T]
}

func (r *_reference[P, T]) IsRef() bool {
	return true
}

func Ref[P kubecrtutils.ObjectPointer[T], T any](name string, target string) Reference {
	return &_reference[P, T]{_definition[P, T]{
		Element: internal.NewElement(name),
		target:  target,
		idxfunc: nil,
		proto:   generics.ObjectFor[P](),
	}}
}

func Define[P kubecrtutils.ObjectPointer[T], T any](name string, target string, idxfunc IndexerFunc[P]) TypedDefinition[P, T] {
	return &_definition[P, T]{
		Element: internal.NewElement(name),
		target:  target,
		idxfunc: idxfunc,
		proto:   generics.ObjectFor[P](),
	}
}

func (d *_definition[P, T]) GetTarget() string {
	return d.target
}

func (d *_definition[P, T]) GetResource() client.Object {
	return d.proto
}

func (d *_definition[P, T]) GetIndexerFunc() client.IndexerFunc {
	return d.indexer
}

func (d *_definition[P, T]) indexer(obj client.Object) []string {
	return d.idxfunc(obj.(any).(P))

}

func (d *_definition[P, T]) Apply(ctx context.Context, set Clusters, logger logging.Logger) (types.Index, error) {
	c := set.Get(d.GetTarget())
	if c == nil {
		return nil, fmt.Errorf("cluster %q not found", d.GetTarget())
	}

	gk, err := kubecrtutils.GKForObject(c, d.proto)
	if err != nil {
		return nil, fmt.Errorf("cannot determine group/kind for %T: %w", d.proto, err)
	}

	i := func(obj client.Object) []string {
		r := d.idxfunc(obj.(any).(P))
		if len(r) > 0 {
			logger.Info("indexing {{key}}: {{values}}", "key", client.ObjectKeyFromObject(obj), "values", fmt.Sprintf("%+v", r))
		}
		return r
	}

	logger.Info("creating index {{index}} for {{resource}} on {{cluster}}[{{effcluster}}]", "index", d.GetName(), "resource", gk, "cluster", d.GetTarget(), "effcluster", c.GetEffective().GetName())
	idx, err := c.CreateIndex(ctx, d.GetName(), d.proto, i, func(_c ClusterEquivalent, name string) (Index, error) {
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

func (d *_definition[P, T]) TypedApply(ctx context.Context, set Clusters, logger logging.Logger) (TypedIndex[T], error) {
	idx, err := d.Apply(ctx, set, logger)
	if err != nil {
		return nil, err
	}
	return idx.(TypedIndex[T]), nil
}
