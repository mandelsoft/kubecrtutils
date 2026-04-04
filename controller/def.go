package controller

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/cluster/clustercontext"
	"github.com/mandelsoft/kubecrtutils/component"
	"github.com/mandelsoft/kubecrtutils/controller/builder"
	"github.com/mandelsoft/kubecrtutils/controller/constraints"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/owner"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/mandelsoft/logging"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ReconcilerFactory is responsible to create a reconciler for the
// given object type.
// Optionally it might implement ModifyFinalizer to influence the finalizer
// name generation.
type ReconcilerFactory[P kubecrtutils.ObjectPointer[T], T any] interface {
	CreateReconciler(ctx context.Context, controller TypedController[P, T], b builder.Builder) (reconcile.Reconciler, error)
}

type ReconcilerFactoryFunc[P kubecrtutils.ObjectPointer[T], T any] func(ctx context.Context, controller TypedController[P, T], b builder.Builder) (reconcile.Reconciler, error)

func (f ReconcilerFactoryFunc[P, T]) CreateReconciler(ctx context.Context, controller TypedController[P, T], b builder.Builder) (reconcile.Reconciler, error) {
	return f(ctx, controller, b)
}

////////////////////////////////////////////////////////////////////////////////

type TypedDefinition[P kubecrtutils.ObjectPointer[T], T any] interface {
	Definition

	GetReconciler() ReconcilerFactory[P, T]
	GetTriggers() []ResourceTriggerDefinition

	WithFinalizer(string) TypedDefinition[P, T]
	WithPredicates(preds ...predicate.Predicate) TypedDefinition[P, T]
	// WithActivationConstraint declares additional activation rules
	// relevant if this controller is activated.
	WithActivationConstraint(...constraints.Constraint) TypedDefinition[P, T]
	InGroup(...string) TypedDefinition[P, T]

	AddForeignIndex(i ...cacheindex.Definition) TypedDefinition[P, T]
	AddIndex(name string, indexerFunc cacheindex.IndexerFunc[P]) TypedDefinition[P, T]
	ImportIndex(reference cacheindex.Reference) TypedDefinition[P, T]
	AddTrigger(trigger ...ResourceTriggerDefinition) TypedDefinition[P, T]
	UseCluster(name ...string) TypedDefinition[P, T]
	UseComponent(name ...string) TypedDefinition[P, T]
}

type _definition[P kubecrtutils.ObjectPointer[T], T any] struct {
	internal.Element
	internal.ErrorContainer
	predicates  []predicate.Predicate
	cluster     string
	clusters    ClusterNames
	components  component.ComponentNames
	proto       client.Object
	reconciler  ReconcilerFactory[P, T]
	indices     map[string]cacheindex.TypedDefinition[P, T]
	imports     map[string]cacheindex.Definition
	triggers    []ResourceTriggerDefinition
	constraints constraints.Constraints
	groups      set.Set[string]
	finalizer   string

	foreign cacheindex.Definitions
}

func DefineByFunc[P kubecrtutils.ObjectPointer[T], T any](name string, cluster string, fac ReconcilerFactoryFunc[P, T]) TypedDefinition[P, T] {
	return Define[P, T](name, cluster, fac)
}

func Define[P kubecrtutils.ObjectPointer[T], T any](name string, cluster string, fac ReconcilerFactory[P, T]) TypedDefinition[P, T] {
	d := &_definition[P, T]{
		Element:        internal.NewElement(name),
		ErrorContainer: *internal.NewErrorContainer(fmt.Sprintf("controller %s", name)),
		cluster:        cluster,
		components:     component.ComponentNames{},
		clusters:       ClusterNames{},
		proto:          generics.ObjectFor[P](),
		reconciler:     fac,
		indices:        map[string]cacheindex.TypedDefinition[P, T]{},
		imports:        map[string]cacheindex.Definition{},
		groups:         set.New[string](),
		constraints:    constraints.New(),
		foreign:        cacheindex.NewDefinitions(),
	}
	d.clusters.Add(cluster)
	return d
}

func (d *_definition[P, T]) InGroup(group ...string) TypedDefinition[P, T] {
	d.groups.Add(group...)
	return d
}

func (d *_definition[P, T]) WithPredicates(preds ...predicate.Predicate) TypedDefinition[P, T] {
	d.predicates = append(d.predicates, preds...)
	return d
}

func (d *_definition[P, T]) WithFinalizer(s string) TypedDefinition[P, T] {
	d.finalizer = s
	return d
}

func (d *_definition[P, T]) UseCluster(name ...string) TypedDefinition[P, T] {
	d.clusters.Add(name...)
	return d
}

func (d *_definition[P, T]) UseComponent(name ...string) TypedDefinition[P, T] {
	d.components.Add(name...)
	return d
}

func (d *_definition[P, T]) WithActivationConstraint(constraints ...constraints.Constraint) TypedDefinition[P, T] {
	d.constraints.Add(constraints...)
	return d
}

func (d *_definition[P, T]) AddForeignIndex(indices ...cacheindex.Definition) TypedDefinition[P, T] {
	for _, i := range indices {
		name := i.GetName()
		if d.indices[name] != nil || d.imports[name] != nil || d.foreign.Get(i.GetName()) != nil {
			d.AddError(fmt.Errorf("duplicate definition of index %q", name))
		} else {
			d.foreign.Add(i)
		}
	}
	return d
}

func (d *_definition[P, T]) AddIndex(name string, indexerFunc cacheindex.IndexerFunc[P]) TypedDefinition[P, T] {
	if d.indices[name] != nil || d.imports[name] != nil || d.foreign.Get(name) != nil {
		d.AddError(fmt.Errorf("duplicate definition of index %q", name))
	} else {
		i := cacheindex.Define[P, T](name, d.cluster, indexerFunc)
		d.indices[name] = i
		d.AddError(i, "index ", name)
		d.clusters.Add(i.GetTarget())
	}
	return d
}

func (d *_definition[P, T]) ImportIndex(def cacheindex.Reference) TypedDefinition[P, T] {
	if d.indices[def.GetName()] != nil || d.imports[def.GetName()] != nil || d.foreign.Get(def.GetName()) != nil {
		d.AddError(fmt.Errorf("duplicate dedinition of index %q", def.GetName()))
	} else {
		d.AddError(def, "index ", def.GetName())
		d.imports[def.GetName()] = def
	}
	return d
}

func (d *_definition[P, T]) AddTrigger(trigger ...ResourceTriggerDefinition) TypedDefinition[P, T] {
	for _, t := range trigger {
		d.triggers = append(d.triggers, t)
		if t.GetCluster() != "" {
			d.clusters.Add(t.GetCluster())
		}
	}
	return d
}

////////////////////////////////////////////////////////////////////////////////

func (d *_definition[P, T]) AddFlags(fs *pflag.FlagSet) {
	if o, ok := d.reconciler.(flagutils.Options); ok {
		o.AddFlags(fs)
	}
}

func (d *_definition[P, T]) AsOptionSet() flagutils.OptionSet {
	if o, ok := d.reconciler.(flagutils.OptionSetProvider); ok {
		return o.AsOptionSet()
	}
	if o, ok := d.reconciler.(flagutils.Options); ok {
		return flagutils.NewOptionSet(o)
	}
	return flagutils.NewOptionSet()
}

func (d *_definition[P, T]) Validate(ctx context.Context, opts flagutils.OptionSet, v flagutils.ValidationSet) error {
	if o, ok := d.reconciler.(flagutils.Validatable); ok {
		return o.Validate(ctx, opts, v)
	}
	for _, i := range d.foreign.Elements {
		if !d.clusters.Contains(i.GetTarget()) {
			return fmt.Errorf("foreign index %q uses undeclared cluster %q", i.GetName(), i.GetTarget())
		}
	}
	return v.ValidateSet(ctx, opts, d.AsOptionSet())
}

func (d *_definition[P, T]) Finalize(ctx context.Context, opts flagutils.OptionSet, v flagutils.FinalizationSet) error {
	if o, ok := d.reconciler.(flagutils.Finalizable); ok {
		return o.Finalize(ctx, opts, v)
	}
	return v.FinalizeSet(ctx, opts, d.AsOptionSet())
}

////////////////////////////////////////////////////////////////////////////////

func (d *_definition[P, T]) GetCluster() string {
	return d.cluster
}

func (d *_definition[P, T]) GetFinalizer() string {
	if d.finalizer == "" {
		return d.GetName()
	}
	return d.finalizer
}

func (d *_definition[P, T]) GetComponents() component.ComponentNames {
	return maps.Clone(d.components)
}

func (d *_definition[P, T]) GetClusters() ClusterNames {
	return maps.Clone(d.clusters)
}

func (d *_definition[P, T]) GetActivationConstraints() constraints.Constraints {
	return d.constraints.Clone()
}

func (d *_definition[P, T]) GetRequiredClusters(mappings mapping.ControllerMappings) ClusterNames {
	names := set.New[string]()
	m := mapping.DefaultMappings(mappings).ClusterMappings()
	for n := range d.GetClusters() {
		names.Add(m.Map(n))
	}
	return names
}

func (d *_definition[P, T]) GetRequiredComponents(mappings mapping.ControllerMappings) component.ComponentNames {
	names := set.New[string]()
	m := mapping.DefaultMappings(mappings).ComponentMappings()
	for n := range d.GetComponents() {
		names.Add(m.Map(n))
	}
	return names
}

func (d *_definition[T, P]) GetGroups() set.Set[string] {
	return maps.Clone(d.groups)
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

func (d *_definition[P, T]) GetForeignIndices() cacheindex.Definitions {
	return d.foreign
}

func (d *_definition[P, T]) CreateIndices(ctx context.Context, m mapping.ControllerMappings, mgr types.ControllerManager) error {
	m = mapping.DefaultMappings(m)
	clusters, err := cluster.Map(mgr.GetClusters(), m.ClusterMappings(), d.GetClusters())
	if err != nil {
		return err
	}
	logger := mgr.GetLogger().WithName(d.GetName()).WithValues("controller", d.GetName())
	idxmap := m.IndexMappings()
	for n, i := range d.indices {
		g := idxmap.Map(n)
		logger.Info("- configuring index {{index}}[{{global}}] from controller {{controller}}", "index", n, "global", g, "controller", d.GetName())
		idx, err := i.Apply(ctx, clusters, logger)
		if err != nil {
			return fmt.Errorf("index %q[%s]: %w", n, g, err)
		}
		logger.Info("  exporting index {{index}}[{{global}}}", "index", n, "global", g)
		err = mgr.GetIndices().Add(cacheindex.NewAlias(g, idx))
		if err != nil {
			return fmt.Errorf("global index %q[%s]: %w", n, g, err)
		}
	}

	// hmm we could add the foreign indices directly to the global index defintions.
	// and use here simple imports instead.

	for n, i := range d.foreign.Elements {
		g := idxmap.Map(n)
		logger.Info("- configuring foreign index {{index}}[{{global}}] from controller {{controller}}", "index", n, "global", g, "controller", d.GetName())
		idx, err := i.Apply(ctx, clusters, logger)
		if err != nil {
			return fmt.Errorf("index %q[%s]: %w", n, g, err)
		}
		logger.Info("  exporting foreign index {{index}}[{{global}}}", "index", n, "global", g)
		err = mgr.GetIndices().Add(cacheindex.NewAlias(g, idx))
		if err != nil {
			return fmt.Errorf("global index %q[%s]: %w", n, g, err)
		}
	}
	return nil
}

func registerIndex[I cacheindex.Index](logger logging.Logger, i cacheindex.Definition, clusters types.Clusters, idxmap mapping.Mappings, mgr types.ControllerManager, local map[string]I) (cacheindex.Index, error) {
	n := i.GetName()
	g := idxmap.Map(n)
	// import indexer
	idx := mgr.GetIndex(g)
	if idx == nil {
		return nil, fmt.Errorf("imported index %q[%s] not found", n, g)
	}

	f := i.GetIndexer()
	if f == nil {
		logger.Info("  importing index {{index}}[{{global}}}", "index", n, "global", g)
	} else {
		logger.Info("  using local index {{index}}[{{global}}}", "index", n, "global", g)
	}
	if reflect.TypeOf(i.GetResource()) != reflect.TypeOf(idx.GetResource()) {
		return nil, fmt.Errorf("index %q resource type mismatch: expected %T, but found %T", n, i.GetResource(), idx.GetResource())
	}
	c := clusters.Get(i.GetTarget())
	if c.GetEffective() != idx.GetCluster().GetEffective() {
		return nil, fmt.Errorf("index %q cluster mismatch: expected %s[%s], but found %s", n, i.GetTarget(), c.GetEffective().GetName(), idx.GetCluster().GetEffective().GetName())
	}
	local[n] = generics.Cast[I](idx.GetEffective())
	return idx.GetEffective(), nil
}

func (d *_definition[P, T]) Apply(ctx context.Context, m mapping.ControllerMappings, mgr types.ControllerManager) (types.Controller, error) {
	if d.GetError() != nil {
		return nil, d.GetError()
	}
	m = mapping.DefaultMappings(m)
	logger := mgr.GetLogger().WithName(d.GetName()).WithValues("controller", d.GetName())
	logger.Info("- configure controller {{controller}}")

	comps, err := component.Map(mgr.GetComponents(), m.ComponentMappings(), d.GetComponents())
	if err != nil {
		return nil, err
	}

	clusters, err := cluster.Map(mgr.GetClusters(), m.ClusterMappings(), d.GetClusters())
	if err != nil {
		return nil, err
	}

	// keep logical view on technical cluster as requested by the definition
	c := clusters.Get(d.cluster)
	if c == nil {
		return nil, fmt.Errorf("cluster %q not found", d.cluster)
	}

	gk, err := kubecrtutils.GKForObject(c, d.proto)
	if err != nil {
		return nil, fmt.Errorf("main resource: %w", err)
	}

	local := map[string]cacheindex.TypedIndex[T]{}
	all := map[string]cacheindex.Index{}
	idxmap := m.IndexMappings()
	for n, i := range d.indices {
		idx, err := registerIndex(logger, i, clusters, idxmap, mgr, local)
		if err != nil {
			return nil, err
		}
		all[n] = idx
	}
	for _, i := range d.foreign.Elements {
		_, err := registerIndex(logger, i, clusters, idxmap, mgr, all)
		if err != nil {
			return nil, err
		}
	}
	for _, i := range d.imports {
		_, err := registerIndex(logger, i, clusters, idxmap, mgr, all)
		if err != nil {
			return nil, err
		}
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

	finalizer := mgr.GetName() + "/" + d.GetFinalizer()
	if m, ok := d.GetReconciler().(FinalizerModifier); ok {
		finalizer = m.ModifyFinalizer(finalizer)
	}
	controller := &_controller[P, T]{
		controllerManager: mgr,
		logger:            logger,
		mappings:          m.ClusterMappings(),
		components:        comps,
		clusters:          clusters,
		cluster:           c,
		gk:                gk,
		definition:        d,
		recorder:          f,
		localIndices:      local,
		allIndices:        all,
		ohandler:          owner.NewHandler(c),
		finalizer:         finalizer,
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
