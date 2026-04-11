package cacheindex

import (
	"context"
	"fmt"

	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/objutils"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reference interface {
	Definition
	IsRef() bool
}

type TypedDefinition[P kubecrtutils.ObjectPointer[T], T any] interface {
	Definition
}

type _definition[P kubecrtutils.ObjectPointer[T], T any] struct {
	internal.Element
	mapping.DefaultClusterConsumer
	target         string
	proto          client.Object
	indexerFactory IndexerFactory
}

type _reference[P kubecrtutils.ObjectPointer[T], T any] struct {
	_definition[P, T]
}

func (r *_reference[P, T]) IsRef() bool {
	return true
}

func Ref[P kubecrtutils.ObjectPointer[T], T any](name string, target string) Reference {
	return &_reference[P, T]{_definition[P, T]{
		Element:                internal.NewElement(ComposeName(name, target)),
		DefaultClusterConsumer: *mapping.NewDefaultClusterConsumer(target),
		target:                 target,
		indexerFactory:         nil,
		proto:                  generics.ObjectFor[P](),
	}}
}

func Define[P kubecrtutils.ObjectPointer[T], T any](name string, target string, idxfunc IndexerFunc[P]) TypedDefinition[P, T] {
	return &_definition[P, T]{
		Element:                internal.NewElement(ComposeName(name, target)),
		DefaultClusterConsumer: *mapping.NewDefaultClusterConsumer(target),
		target:                 target,
		indexerFactory:         Lift(ConvertIndexerFunc(idxfunc)),
		proto:                  generics.ObjectFor[P](),
	}
}

func DefineByFactory[P kubecrtutils.ObjectPointer[T], T any](name string, target string, idxfactory TypedIndexerFactory[P]) TypedDefinition[P, T] {
	return &_definition[P, T]{
		Element:                internal.NewElement(ComposeName(name, target)),
		DefaultClusterConsumer: *mapping.NewDefaultClusterConsumer(target),
		target:                 target,
		indexerFactory:         ConvertIndexerFactory(idxfactory),
		proto:                  generics.ObjectFor[P](),
	}
}

func (d *_definition[P, T]) GetEffective() Definition {
	return d
}

func (d *_definition[P, T]) GetTarget() string {
	return d.target
}

func (d *_definition[P, T]) GetResource() client.Object {
	return d.proto
}

func (d *_definition[P, T]) GetIndexer() IndexerFactory {
	return d.indexerFactory
}

func (d *_definition[P, T]) Apply(ctx context.Context, mappings mapping.ControllerMappings, mgr types.ControllerManager) error {
	logger := mgr.GetLogger()
	mappings = mapping.DefaultMappings(mappings)
	clusters, err := mgr.GetClusters().Map(mappings.ClusterMappings(), d.GetClusters())
	if err != nil {
		return err
	}
	c := clusters.Get(d.target).GetEffective()
	glob := MapName(d.GetName(), mappings)

	if c == nil {
		return fmt.Errorf("cluster %s->%s not found", d.GetTarget(), c.GetName())
	}

	gk, err := objutils.GKForObject(c, d.proto)
	if err != nil {
		return fmt.Errorf("cannot determine group/kind for %T: %w", d.proto, err)
	}
	if d.GetIndexer() == nil {
		return fmt.Errorf("indexer required for %T", d.proto)
	}

	old := mgr.GetIndex(glob)
	if old != nil {
		if err := Match(old, gk, c); err != nil {
			return err
		}
		logger.Info("  sharing existing global index {{index}} on resource {{resource}} for {{local}}->{{global}} on {{cluster}}->{{effective}}", "index", glob, "local", d.GetName(), "cluster", d.GetTarget(), "effective", c.GetName(), "resource", gk)
		return nil
	}
	logger.Info("  creating index {{local}}->{{index}} on resource {{resource}}  on {{cluster}}->{{effective}}", "index", glob, "local", d.GetName(), "cluster", d.GetTarget(), "effective", c.GetName(), "resource", gk)

	ilog := logger.WithName("index."+glob).WithValues("index", glob)
	f, err := d.indexerFactory(ctx, ilog, clusters)
	if err != nil {
		return err
	}
	i := func(obj client.Object) []string {
		r := f(obj)
		if len(r) > 0 {
			ilog.Info("indexing {{key}}: {{values}}", "key", client.ObjectKeyFromObject(obj), "values", fmt.Sprintf("%+v", r))
		}
		return r
	}

	idx, err := c.CreateIndex(ctx, glob, d.proto, i, func(_c ClusterEquivalent, name string) (Index, error) {
		idx, err := NewDefaultIndex(name, _c, d.proto)
		if err != nil {
			return nil, err
		}
		return &_typedIndex[T]{
			idx,
		}, nil
	})

	if err != nil {
		return err
	}

	return mgr.GetIndices().Add(idx)
}

func Match(i Index, gk schema.GroupKind, c types.ClusterEquivalent) error {
	if i.GetGVK().GroupKind() != gk {
		return fmt.Errorf("group kind mismatch: expected %s, but found %s", gk, i.GetGVK())
	}
	if i.GetCluster().GetEffective() != c.GetEffective() {
		return fmt.Errorf("cluster mismatch: expected %s->%s, but found %s", c.GetName(), c.GetEffective().GetName(), i.GetCluster().GetEffective().GetName())
	}
	return nil
}
