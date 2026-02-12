package controller

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/cluster/clustercontext"
	"github.com/mandelsoft/kubecrtutils/controller/builder"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/owner"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReconcilerFactory[P kubecrtutils.ObjectPointer[T], T any] interface {
	CreateReconciler(ctx context.Context, controller Controller[P, T], b builder.Builder) (reconcile.Reconciler, error)
}

type ReconcilerFactoryFunc[P kubecrtutils.ObjectPointer[T], T any] func(ctx context.Context, controller Controller[P, T], b builder.Builder) (reconcile.Reconciler, error)

func (f ReconcilerFactoryFunc[P, T]) CreateReconciler(ctx context.Context, controller Controller[P, T], b builder.Builder) (reconcile.Reconciler, error) {
	return f(ctx, controller, b)
}

////////////////////////////////////////////////////////////////////////////////

type Definition = types.ControllerDefinition

type TypedDefinition[P kubecrtutils.ObjectPointer[T], T any] interface {
	Definition

	GetReconciler() ReconcilerFactory[P, T]
	GetTriggers() []ResourceTriggerDefinition

	AddIndex(name string, indexerFunc cacheindex.IndexerFunc[P]) TypedDefinition[P, T]
	AddTrigger(trigger ...ResourceTriggerDefinition) TypedDefinition[P, T]
	UseCluster(name ...string) TypedDefinition[P, T]
}

type _definition[P kubecrtutils.ObjectPointer[T], T any] struct {
	internal.Element
	predicates []predicate.Predicate
	cluster    string
	clusters   sets.Set[string]
	proto      client.Object
	reconciler ReconcilerFactory[P, T]
	indices    map[string]cacheindex.TypedDefinition[P, T]
	triggers   []ResourceTriggerDefinition
	err        error
}

func DefineByFunc[P kubecrtutils.ObjectPointer[T], T any](name string, cluster string, fac ReconcilerFactoryFunc[P, T]) TypedDefinition[P, T] {
	return Define[P, T](name, cluster, fac)
}

func Define[P kubecrtutils.ObjectPointer[T], T any](name string, cluster string, fac ReconcilerFactory[P, T]) TypedDefinition[P, T] {
	d := &_definition[P, T]{
		Element:    internal.NewElement(name),
		cluster:    cluster,
		clusters:   sets.New[string](cluster),
		proto:      generics.ObjectFor[P](),
		reconciler: fac,
		indices:    map[string]cacheindex.TypedDefinition[P, T]{},
	}
	return d
}

func (d *_definition[P, T]) WithPredicates(preds ...predicate.Predicate) *_definition[P, T] {
	d.predicates = append(d.predicates, preds...)
	return d
}

func (d *_definition[P, T]) UseCluster(name ...string) TypedDefinition[P, T] {
	d.clusters.Insert(name...)
	return d
}

func (d *_definition[P, T]) AddIndex(name string, indexerFunc cacheindex.IndexerFunc[P]) TypedDefinition[P, T] {
	if d.indices[name] != nil {
		d.err = errors.Join(d.err, fmt.Errorf("duplicate deinition of index %q", name))
	} else {
		d.indices[name] = cacheindex.NewDefinition[P, T](GlobalControllerIndexName(d.GetName(), name), d.cluster, indexerFunc)
	}
	return d
}

func (d *_definition[P, T]) AddTrigger(trigger ...ResourceTriggerDefinition) TypedDefinition[P, T] {
	for _, t := range trigger {
		d.triggers = append(d.triggers, t)
		if t.GetCluster() != "" {
			d.clusters.Insert(t.GetCluster())
		}
	}
	return d
}

func (d *_definition[P, T]) AddFlags(fs *pflag.FlagSet) {
	if o, ok := d.reconciler.(flagutils.Options); ok {
		o.AddFlags(fs)
	}
}

func (d *_definition[P, T]) AsOptionSet() flagutils.OptionSet {
	if o, ok := d.reconciler.(flagutils.OptionSetProvider); ok {
		return o.AsOptionSet()
	}
	return flagutils.DefaultOptionSet{}
}

func (d *_definition[P, T]) Validate(ctx context.Context, opts flagutils.OptionSet, v flagutils.ValidationSet) error {
	if o, ok := d.reconciler.(flagutils.Validatable); ok {
		return o.Validate(ctx, opts, v)
	}
	return v.ValidateSet(ctx, opts, d.AsOptionSet())
}

func (d *_definition[P, T]) Finalize(ctx context.Context, opts flagutils.OptionSet, v flagutils.FinalizationSet) error {
	if o, ok := d.reconciler.(flagutils.Finalizable); ok {
		return o.Finalize(ctx, opts, v)
	}
	return v.FinalizeSet(ctx, opts, d.AsOptionSet())
}

func (d *_definition[P, T]) GetError() error {
	return d.err
}

func (d *_definition[P, T]) GetCluster() string {
	return d.cluster
}

func (d *_definition[P, T]) GetClusters() sets.Set[string] {
	return maps.Clone(d.clusters)
}

func (d *_definition[T, P]) GetResource() client.Object {
	return d.proto
}

func (d *_definition[P, T]) GetWatchPredicates() []predicate.Predicate {
	return slices.Clone(d.predicates)
}

func (d *_definition[P, T]) GetReconciler() ReconcilerFactory[P, T] {
	return d.reconciler
}

func (d *_definition[P, T]) GetTriggers() []ResourceTriggerDefinition {
	return slices.Clone(d.triggers)
}

func (d *_definition[P, T]) CreateController(ctx context.Context, mgr types.ControllerManager) (types.Controller, error) {
	logger := mgr.GetLogger().WithName(d.GetName()).WithValues("controller", d.GetName())
	logger.Info("configure controller {{controller}}")

	// TODO: make controller composable
	mapping := cluster.IdMapping(d.GetClusters())
	clusters, err := cluster.Map(mgr.GetClusters(), mapping)
	if err != nil {
		return nil, err
	}

	c := clusters.Get(d.cluster)
	if c == nil {
		return nil, fmt.Errorf("cluster %q not found", d.GetName())
	}

	gk, err := kubecrtutils.GKForObject(c, d.proto)
	if err != nil {
		return nil, fmt.Errorf("main resource: %w", err)
	}

	local := map[string]cacheindex.TypedIndex[T]{}
	for n, i := range d.indices {
		idx, err := i.Apply(ctx, clusters, logger)
		if err != nil {
			return nil, fmt.Errorf("index %q[%s]: %w", n, gk, err)
		}
		mgr.GetIndices().Add(idx)
		local[n] = idx.(cacheindex.TypedIndex[T])
	}

	evtSource := mgr.GetName() + "/" + c.GetName()
	var f recorderFunc

	if c.AsFleet() != nil {
		f = func(ctx context.Context) record.EventRecorder {
			return clustercontext.ClusterFor(ctx).GetEventRecorderFor(evtSource)
		}
	} else {
		s := c.AsCluster().GetEventRecorderFor(evtSource)
		f = func(ctx context.Context) record.EventRecorder {
			return s
		}
	}
	controller := &_controller[P, T]{
		controllerManager: mgr,
		logger:            logger,
		clusters:          clusters, // TODO; name mapping
		cluster:           c,
		gk:                gk,
		definition:        d,
		recorder:          f,
		indices:           local,
		ohandler:          owner.NewHandler(c),
	}
	return controller, nil
}

func (d *_definition[P, T]) GetOptions() flagutils.Options {
	if o, ok := d.reconciler.(flagutils.Options); ok {
		return o
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func GlobalControllerIndexName(cname, iname string) string {
	return cname + ":" + iname
}
