package controller

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/cacheindex/idxutils"
	"github.com/mandelsoft/kubecrtutils/cluster/clustercontext"
	"github.com/mandelsoft/kubecrtutils/controller/builder"
	"github.com/mandelsoft/kubecrtutils/controller/constraints"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/objutils"
	"github.com/mandelsoft/kubecrtutils/owner"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func DefinitionFromContext(ctx context.Context) Definition {
	return generics.Cast[Definition](ctx.Value("controller"))
}

// --- begin reconciler factory ---
// ReconcilerFactory is responsible to create a reconciler for the
// given object type.
// Optionally it might implement ModifyFinalizer to influence the finalizer
// name generation.
type ReconcilerFactory[P kubecrtutils.ObjectPointer[T], T any] interface {
	CreateReconciler(ctx context.Context, controller TypedController[P, T], b builder.Builder) (reconcile.Reconciler, error)
}

// --- end reconciler factory ---

// --- begin reconciler factory function ---

type ReconcilerFactoryFunc[P kubecrtutils.ObjectPointer[T], T any] func(ctx context.Context, controller TypedController[P, T], b builder.Builder) (reconcile.Reconciler, error)

// --- end reconciler factory function ---

func (f ReconcilerFactoryFunc[P, T]) CreateReconciler(ctx context.Context, controller TypedController[P, T], b builder.Builder) (reconcile.Reconciler, error) {
	return f(ctx, controller, b)
}

type IndexerFactory[T client.Object] = cacheindex.TypedIndexerFactory[T]

////////////////////////////////////////////////////////////////////////////////

// --- begin definition ---

type CompositionInterface[P kubecrtutils.ObjectPointer[T], T any] interface {
	TypedDefinition[P, T]

	WithFinalizer(string) CompositionInterface[P, T]
	WithPredicates(preds ...predicate.Predicate) CompositionInterface[P, T]
	// WithActivationConstraint declares additional activation rules
	// relevant if this controller is activated.
	WithActivationConstraint(...constraints.Constraint) CompositionInterface[P, T]
	InGroup(...string) CompositionInterface[P, T]

	AddForeignIndex(i ...cacheindex.Definition) CompositionInterface[P, T]
	AddIndex(name string, indexerFunc cacheindex.TypedIndexerFunc[P]) CompositionInterface[P, T]
	AddIndexByFactory(name string, indexerFunc cacheindex.TypedIndexerFactory[P]) CompositionInterface[P, T]
	ImportIndex(reference cacheindex.Reference) CompositionInterface[P, T]
	AddTrigger(trigger ...ResourceTriggerDefinition) CompositionInterface[P, T]
	UseCluster(name ...string) CompositionInterface[P, T]
	UseComponent(name ...string) CompositionInterface[P, T]
}

// --- end definition ---

type TypedDefinition[P kubecrtutils.ObjectPointer[T], T any] interface {
	Definition

	GetReconciler() ReconcilerFactory[P, T]
	GetTriggers() []ResourceTriggerDefinition
}

type _definition[P kubecrtutils.ObjectPointer[T], T any] struct {
	internal.Element
	internal.ErrorContainer
	mapping.DefaultConsumer
	predicates  []predicate.Predicate
	cluster     string
	proto       client.Object
	reconciler  ReconcilerFactory[P, T]
	indices     cacheindex.Definitions
	imports     cacheindex.Definitions
	foreign     cacheindex.Definitions
	triggers    []ResourceTriggerDefinition
	constraints constraints.Constraints
	groups      set.Set[string]

	finalizer string
}

func DefineByFunc[P kubecrtutils.ObjectPointer[T], T any](name string, cluster string, fac ReconcilerFactoryFunc[P, T]) CompositionInterface[P, T] {
	return Define[P, T](name, cluster, fac)
}

func Define[P kubecrtutils.ObjectPointer[T], T any](name string, cluster string, fac ReconcilerFactory[P, T]) CompositionInterface[P, T] {
	d := &_definition[P, T]{
		Element:         internal.NewElement(name),
		ErrorContainer:  *internal.NewErrorContainer(fmt.Sprintf("controller %s", name)),
		DefaultConsumer: *mapping.NewDefaultConsumer(),
		cluster:         cluster,
		proto:           generics.ObjectFor[P](),
		reconciler:      fac,
		indices:         cacheindex.NewDefinitions(),
		foreign:         cacheindex.NewDefinitions(),
		imports:         cacheindex.NewDefinitions(),
		groups:          set.New[string](),
		constraints:     constraints.New(),
	}
	d.UseCluster(cluster)
	return d
}

func (d *_definition[P, T]) InGroup(group ...string) CompositionInterface[P, T] {
	d.groups.Add(group...)
	return d
}

func (d *_definition[P, T]) WithPredicates(preds ...predicate.Predicate) CompositionInterface[P, T] {
	d.predicates = append(d.predicates, preds...)
	return d
}

func (d *_definition[P, T]) WithFinalizer(s string) CompositionInterface[P, T] {
	d.finalizer = s
	return d
}

func (d *_definition[P, T]) UseCluster(name ...string) CompositionInterface[P, T] {
	d.DefaultConsumer.UseCluster(name...)
	return d
}

func (d *_definition[P, T]) UseComponent(name ...string) CompositionInterface[P, T] {
	d.DefaultConsumer.UseComponent(name...)
	return d
}

func (d *_definition[P, T]) WithActivationConstraint(constraints ...constraints.Constraint) CompositionInterface[P, T] {
	d.constraints.Add(constraints...)
	return d
}

func (d *_definition[P, T]) AddForeignIndex(indices ...cacheindex.Definition) CompositionInterface[P, T] {
	for _, i := range indices {
		name := i.GetName()
		if d.indices.Get(name) != nil || d.imports.Get(name) != nil || d.foreign.Get(name) != nil {
			d.AddError(fmt.Errorf("duplicate definition of index %q", name))
		} else {
			d.foreign.Add(i)
		}
	}
	return d
}

func (d *_definition[P, T]) AddIndex(name string, indexerFunc cacheindex.TypedIndexerFunc[P]) CompositionInterface[P, T] {
	n := cacheindex.ComposeName(name, d.cluster)
	if d.indices.Get(n) != nil || d.imports.Get(n) != nil || d.foreign.Get(n) != nil {
		d.AddError(fmt.Errorf("duplicate definition of index %q[%s]", name, n))
	} else {
		i := cacheindex.Define[P, T](name, d.cluster, indexerFunc)
		d.indices.Add(i)
		d.AddError(i, "index ", name)
		d.UseCluster(i.GetTarget())
	}
	return d
}

func (d *_definition[P, T]) AddIndexByFactory(name string, indexerFunc IndexerFactory[P]) CompositionInterface[P, T] {
	n := cacheindex.ComposeName(name, d.cluster)
	if d.indices.Get(n) != nil || d.imports.Get(n) != nil || d.foreign.Get(n) != nil {
		d.AddError(fmt.Errorf("duplicate definition of index %q[%s]", name, n))
	} else {
		i := cacheindex.DefineByFactory[P, T](name, d.cluster, indexerFunc)
		d.indices.Add(i)
		d.AddError(i, "index ", name)
		d.UseCluster(i.GetTarget())
	}
	return d
}

func (d *_definition[P, T]) ImportIndex(def cacheindex.Reference) CompositionInterface[P, T] {
	name := def.GetName()
	if d.indices.Get(name) != nil || d.imports.Get(def.GetName()) != nil || d.foreign.Get(def.GetName()) != nil {
		d.AddError(fmt.Errorf("duplicate dedinition of index %q", def.GetName()))
	} else {
		d.AddError(def, "index ", def.GetName())
		d.imports.Add(def)
	}
	return d
}

func (d *_definition[P, T]) AddTrigger(trigger ...ResourceTriggerDefinition) CompositionInterface[P, T] {
	for _, t := range trigger {
		d.triggers = append(d.triggers, t)
		if t.GetCluster() != "" {
			d.UseCluster(t.GetCluster())
		}
	}
	return d
}

////////////////////////////////////////////////////////////////////////////////

func (d *_definition[P, T]) GetOptions() flagutils.Options {
	if o, ok := d.reconciler.(flagutils.OptionSetProvider); ok {
		return o.AsOptionSet()
	}
	if o, ok := d.reconciler.(flagutils.Options); ok {
		return o
	}
	return nil
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
	if o, ok := d.reconciler.(flagutils.Options); ok {
		return flagutils.NewOptionSet(o)
	}
	return flagutils.NewOptionSet()
}

func (d *_definition[P, T]) Prepare(ctx context.Context, opts flagutils.OptionSet, v flagutils.PreparationSet) error {
	if o, ok := d.reconciler.(flagutils.Preparable); ok {
		return errors.Wrapf(o.Prepare(ctx, opts, v), "%s: ", d.GetName())
	}
	return v.PrepareSet(ctx, opts, d.AsOptionSet())
}

func (d *_definition[P, T]) Validate(ctx context.Context, opts flagutils.OptionSet, v flagutils.ValidationSet) error {
	if o, ok := d.reconciler.(flagutils.Validatable); ok {
		return errors.Wrapf(o.Validate(ctx, opts, v), "%s: ", d.GetName())
	}
	for _, i := range d.foreign.Elements {
		if !d.GetClusters().Contains(i.GetTarget()) {
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

func (d *_definition[P, T]) GetFinalizer() string {
	if d.finalizer == "" {
		return d.GetName()
	}
	return d.finalizer
}

func (d *_definition[P, T]) GetActivationConstraints() constraints.Constraints {
	return d.constraints.Clone()
}

func (d *_definition[P, T]) GetCluster() string {
	return d.cluster
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

func (d *_definition[P, T]) CreateIndices(ctx context.Context, mappings mapping.ControllerMappings, mgr types.ControllerManager) error {
	logger := mgr.GetLogger().WithName(d.GetName()).WithValues("controller", d.GetName())

	ctx = context.WithValue(ctx, "controller", d)
	ctx = context.WithValue(ctx, "options", d.GetOptions())

	for n, i := range d.indices.Elements {
		logger.Info("- configuring local index {{index}} from controller {{controller}}", "index", n, "controller", d.GetName())
		err := i.Apply(ctx, mappings, mgr)
		if err != nil {
			return fmt.Errorf("controller %q: index %q: %w", d.GetName(), n, err)
		}
	}

	for n, i := range d.foreign.Elements {
		logger.Info("- configuring foreign index {{index}} from controller {{controller}}", "index", n, "controller", d.GetName())
		err := i.Apply(ctx, mappings, mgr)
		if err != nil {
			return fmt.Errorf("controller %q: index %q: %w", d.GetName(), n, err)
		}
	}
	return nil
}

func (d *_definition[P, T]) Apply(ctx context.Context, m mapping.ControllerMappings, mgr types.ControllerManager) error {
	if d.GetError() != nil {
		return d.GetError()
	}
	m = mapping.DefaultMappings(m)
	logger := mgr.GetLogger().WithName(d.GetName()).WithValues("controller", d.GetName())
	logger.Info("- configure controller {{controller}}")

	comps, err := mgr.GetComponents().Map(m.ComponentMappings(), d.GetComponents())
	if err != nil {
		return err
	}

	clusters, err := mgr.GetClusters().Map(m.ClusterMappings(), d.GetClusters())
	if err != nil {
		return err
	}

	// keep logical view on technical cluster as requested by the definition
	c := clusters.Get(d.cluster)
	if c == nil {
		return fmt.Errorf("cluster %q not found", d.cluster)
	}

	gk, err := objutils.GKForObject(c, d.proto)
	if err != nil {
		return fmt.Errorf("main resource: %w", err)
	}

	local := map[string]cacheindex.TypedIndex[T]{}
	all := map[string]cacheindex.Index{}
	for n, i := range d.indices.Elements {
		err := idxutils.ImportIndex(logger, i, clusters, m, mgr, local, cacheindex.BaseName)
		if err != nil {
			return err
		}
		all[n] = local[cacheindex.BaseName(n)]
	}

	call, err := idxutils.ImportIndices(all, logger, d.cluster, clusters, m, mgr, d.imports, d.foreign)
	if err != nil {
		return err
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
		indices:           call,
		ohandler:          owner.NewHandler(c),
		finalizer:         finalizer,
	}
	err = mgr.GetControllers().Add(controller)
	if err != nil {
		return err
	}

	return controller.Complete(ctx)
}

////////////////////////////////////////////////////////////////////////////////

func GlobalControllerIndexName(cname, iname string) string {
	return cname + ":" + iname
}
