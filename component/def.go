package component

import (
	"context"
	"fmt"
	"reflect"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/cacheindex/idxutils"
	"github.com/mandelsoft/kubecrtutils/controller/constraints"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/mandelsoft/logging"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func DefinitionFromContext(ctx context.Context) Definition {
	return generics.Cast[Definition](ctx.Value("component"))
}

type Factory interface {
	Apply(ctx context.Context, base *Base) (Component, error)
}

type Definition = interface {
	flagutils.Options
	mapping.Consumer

	GetName() string

	GetForeignIndices() cacheindex.Definitions
	GetActivationConstraints() constraints.Constraints

	GetRequiredClusters(mappings mapping.ControllerMappings) types.ClusterNames
	GetRequiredComponents(mappings mapping.ControllerMappings) ComponentNames

	types.IndexProvider
	types.Applyable

	GetOptions() flagutils.Options
}

type _definition struct {
	internal.Element
	internal.ErrorContainer
	mapping.DefaultConsumer
	constraints constraints.Constraints
	foreign     cacheindex.Definitions
	imports     cacheindex.Definitions

	factory Factory
}

var _ Definition = (*_definition)(nil)

func Define(name string, fac Factory) *_definition {
	return &_definition{
		Element:         internal.NewElement(name),
		ErrorContainer:  *internal.NewErrorContainer(fmt.Sprintf("component %s", name)),
		DefaultConsumer: *mapping.NewDefaultConsumer(),
		constraints:     constraints.New(),
		foreign:         cacheindex.NewDefinitions(),
		imports:         cacheindex.NewDefinitions(),
		factory:         fac,
	}
}

func (d *_definition) UseCluster(name ...string) *_definition {
	d.DefaultConsumer.UseCluster(name...)
	return d
}

func (d *_definition) UseComponent(name ...string) *_definition {
	d.DefaultConsumer.UseComponent(name...)
	return d
}

func (d *_definition) WithActivationConstraint(constraints ...constraints.Constraint) *_definition {
	d.constraints.Add(constraints...)
	return d
}

func (d *_definition) ImportIndex(def cacheindex.Reference) *_definition {
	if d.imports.Get(def.GetName()) != nil || d.foreign.Get(def.GetName()) != nil {
		d.AddError(fmt.Errorf("duplicate dedinition of index %q", def.GetName()))
	} else {
		d.AddError(def, "index ", def.GetName())
		d.imports.Add(def)
	}
	return d
}

func (d *_definition) AddForeignIndex(indices ...cacheindex.Definition) *_definition {
	for _, i := range indices {
		if d.imports.Get(i.GetName()) != nil || d.foreign.Get(i.GetName()) != nil {
			d.AddError(fmt.Errorf("duplicate dedinition of index %q", i.GetName()))
		} else {
			d.foreign.Add(i)
		}
	}
	return d
}

func (d *_definition) GetActivationConstraints() constraints.Constraints {
	return d.constraints.Clone()
}

func (d *_definition) GetForeignIndices() cacheindex.Definitions {
	return d.foreign
}

func (d *_definition) GetOptions() flagutils.Options {
	if o, ok := d.factory.(flagutils.OptionSetProvider); ok {
		return o.AsOptionSet()
	}
	if o, ok := d.factory.(flagutils.Options); ok {
		return o
	}
	return nil
}

func (d *_definition) AddFlags(fs *pflag.FlagSet) {
	if o, ok := d.factory.(flagutils.Options); ok {
		o.AddFlags(fs)
	}
}

func (d *_definition) AsOptionSet() flagutils.OptionSet {
	if o, ok := d.factory.(flagutils.OptionSetProvider); ok {
		return o.AsOptionSet()
	}
	if o, ok := d.factory.(flagutils.Options); ok {
		return flagutils.NewOptionSet(o)
	}
	return flagutils.NewOptionSet()
}

func (d *_definition) Prepare(ctx context.Context, opts flagutils.OptionSet, v flagutils.PreparationSet) error {
	if o, ok := d.factory.(flagutils.Preparable); ok {
		return errors.Wrapf(o.Prepare(ctx, opts, v), "%s: ", d.GetName())
	}
	return v.PrepareSet(ctx, opts, d.AsOptionSet())
}

func (d *_definition) Validate(ctx context.Context, opts flagutils.OptionSet, v flagutils.ValidationSet) error {
	if o, ok := d.factory.(flagutils.Validatable); ok {
		return errors.Wrapf(o.Validate(ctx, opts, v), "%s: ", d.GetName())
	}
	for _, i := range d.foreign.Elements {
		if !d.GetClusters().Contains(i.GetTarget()) {
			return fmt.Errorf("component %q: foreign index %q uses undeclared cluster %q", d.GetName(), i.GetName(), i.GetTarget())
		}
	}
	return v.ValidateSet(ctx, opts, d.AsOptionSet())
}

func (d *_definition) Finalize(ctx context.Context, opts flagutils.OptionSet, v flagutils.FinalizationSet) error {
	if o, ok := d.factory.(flagutils.Finalizable); ok {
		return o.Finalize(ctx, opts, v)
	}
	return v.FinalizeSet(ctx, opts, d.AsOptionSet())
}

func (d *_definition) CreateIndices(ctx context.Context, mappings mapping.ControllerMappings, mgr types.ControllerManager) error {
	logger := mgr.GetLogger().WithName(d.GetName()).WithValues("component", d.GetName())

	ctx = context.WithValue(ctx, "component", d)
	ctx = context.WithValue(ctx, "options", d.GetOptions())
	for n, i := range d.foreign.Elements {
		logger.Info("- configuring foreign index {{index}} from component {{component}}", "index", n, "component", d.GetName())
		err := i.Apply(ctx, mappings, mgr)
		if err != nil {
			return fmt.Errorf("component %q: index %q: %w", d.GetName(), n, err)
		}
	}
	return nil
}

func (d *_definition) Apply(ctx context.Context, m mapping.ControllerMappings, mgr types.ControllerManager) error {
	logger := mgr.GetLogger().WithName(d.GetName()).WithValues("component", d.GetName())
	logger.Info("- configure component {{component}}", "component", d.GetName())

	m = mapping.DefaultMappings(m)
	clusters, err := mgr.GetClusters().Map(m.ClusterMappings(), d.GetClusters())
	if err != nil {
		return err
	}

	comps, err := mgr.GetComponents().Map(m.ComponentMappings(), d.GetComponents())
	if err != nil {
		return err
	}

	all := map[string]cacheindex.Index{}
	err = idxutils.ImportIndices(all, logger, "", clusters, m, mgr, d.foreign, d.imports)
	if err != nil {
		return err
	}

	indices := cacheindex.NewIndices()
	for n, i := range all {
		indices.Add(cacheindex.NewAlias(n, i))
	}

	b := &Base{
		Logger:   logger,
		def:      d,
		self:     nil,
		clusters: clusters,
		comps:    comps,
		indices:  indices,
	}
	c, err := d.factory.Apply(ctx, b)
	if err != nil {
		return err
	}
	b.self = c
	err = mgr.GetComponents().Add(c)
	if err != nil {
		return err
	}
	if r, ok := c.(manager.Runnable); ok {
		logger.Info("  register as runnable")
		mgr.GetManager().GetLocalManager().Add(r)
	}
	return nil
}

func registerIndex[I cacheindex.Index](logger logging.Logger, i cacheindex.Definition, clusters types.Clusters, idxmap mapping.Mappings, mgr types.ControllerManager, local map[string]I) (cacheindex.Index, error) {
	n := i.GetName()
	c := clusters.Get(i.GetTarget()).GetEffective()
	glob := cacheindex.ComposeName(idxmap.Map(n), c.GetName())

	// import indexer
	idx := mgr.GetIndex(glob)
	if idx == nil {
		return nil, fmt.Errorf("index %q->%q not found", n, glob)
	}

	f := i.GetIndexer()
	if f == nil {
		logger.Info("  importing index {{index}}->{{global}}", "index", n, "global", glob)
	} else {
		logger.Info("  using local index {{index}}->{{global}}", "index", n, "global", glob)
	}
	if reflect.TypeOf(i.GetResource()) != reflect.TypeOf(idx.GetResource()) {
		return nil, fmt.Errorf("index %q->%q resource type mismatch: expected %T, but found %T", n, glob, i.GetResource(), idx.GetResource())
	}
	if c.GetEffective() != idx.GetCluster().GetEffective() {
		return nil, fmt.Errorf("index %q->%q cluster mismatch: expected %s[%s], but found %s", n, glob, i.GetTarget(), c.GetEffective().GetName(), idx.GetCluster().GetEffective().GetName())
	}
	local[n] = generics.Cast[I](idx.GetEffective())
	return idx.GetEffective(), nil
}
