package component

import (
	"context"
	"fmt"
	"maps"
	"reflect"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/maputils"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/controller/constraints"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/mandelsoft/logging"
	"github.com/spf13/pflag"
)

type Factory interface {
	Apply(ctx context.Context, def Definition, set types.Clusters, indices cacheindex.Indices, logger logging.Logger) (types.Index, error)
}

type Definition = interface {
	flagutils.Options
	GetName() string

	GetForeignIndices() cacheindex.Definitions
	GetActivationConstraints() constraints.Constraints
	GetRequiredClusters(mappings types.ControllerMappings) types.ClusterNames

	CreateIndices(ctx context.Context, mapping types.ControllerMappings, mgr types.ControllerManager) error
	CreateComponent(ctx context.Context, mapping types.ControllerMappings, mgr types.ControllerManager) (Component, error)
}

type _definition struct {
	internal.Element
	internal.ErrorContainer
	clusters    types.ClusterNames
	constraints constraints.Constraints
	foreign     cacheindex.Definitions
	imports     map[string]cacheindex.Definition

	factory Factory
}

func Define(name string, fac Factory) Definition {
	return &_definition{
		Element:        internal.NewElement(name),
		ErrorContainer: *internal.NewErrorContainer(fmt.Sprintf("component %s", name)),
		clusters:       types.ClusterNames{},
		constraints:    constraints.New(),
		foreign:        cacheindex.NewDefinitions(),
		imports:        map[string]cacheindex.Definition{},
		factory:        fac,
	}
}

func (d *_definition) UseCluster(name ...string) *_definition {
	d.clusters.Add(name...)
	return d
}

func (d *_definition) WithActivationConstraint(constraints ...constraints.Constraint) *_definition {
	d.constraints.Add(constraints...)
	return d
}

func (d *_definition) ImportIndex(def cacheindex.Reference) *_definition {
	if d.imports[def.GetName()] != nil || d.foreign.Get(def.GetName()) != nil {
		d.AddError(fmt.Errorf("duplicate dedinition of index %q", def.GetName()))
	} else {
		d.AddError(def, "index ", def.GetName())
		d.imports[def.GetName()] = def
	}
	return d
}

func (d *_definition) AddForeignIndex(indices ...cacheindex.Definition) *_definition {
	for _, i := range indices {
		if d.imports[i.GetName()] != nil || d.foreign.Get(i.GetName()) != nil {
			d.AddError(fmt.Errorf("duplicate dedinition of index %q", i.GetName()))
		} else {
			d.foreign.Add(i)
		}
	}
	return d
}

func (d *_definition) GetClusters() types.ClusterNames {
	return maps.Clone(d.clusters)
}

func (d *_definition) GetActivationConstraints() constraints.Constraints {
	return d.constraints.Clone()
}

func (d *_definition) GetRequiredClusters(mappings types.ControllerMappings) types.ClusterNames {
	names := set.New[string]()
	m := types.DefaultMappings(mappings).ClusterMappings()
	for n := range d.GetClusters() {
		names.Add(m.Map(n))
	}
	return names
}

func (d *_definition) GetForeignIndices() cacheindex.Definitions {
	return d.foreign
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

func (d *_definition) Validate(ctx context.Context, opts flagutils.OptionSet, v flagutils.ValidationSet) error {
	if o, ok := d.factory.(flagutils.Validatable); ok {
		return o.Validate(ctx, opts, v)
	}
	for _, i := range d.foreign.Elements {
		if !d.clusters.Contains(i.GetTarget()) {
			return fmt.Errorf("foreign index %q usines undeclared cluster %q", i.GetName(), i.GetTarget())
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

func (d *_definition) CreateIndices(ctx context.Context, mapping types.ControllerMappings, mgr types.ControllerManager) error {
	mapping = types.DefaultMappings(mapping)
	clusters, err := cluster.Map(mgr.GetClusters(), mapping.ClusterMappings(), d.GetClusters())
	if err != nil {
		return err
	}
	logger := mgr.GetLogger().WithName(d.GetName()).WithValues("component", d.GetName())
	idxmap := mapping.IndexMappings()

	// hmm we could add the foreign indices directly to the global index definitions.
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

func (d *_definition) CreateComponent(ctx context.Context, mapping types.ControllerMappings, mgr types.ControllerManager) (Component, error) {
	logger := mgr.GetLogger().WithName(d.GetName()).WithValues("component", d.GetName())
	logger.Info("- configure controller {{controller}}")

	clusters, err := cluster.Map(mgr.GetClusters(), mapping.ClusterMappings(), d.GetClusters())
	if err != nil {
		return nil, err
	}

	all := map[string]cacheindex.Index{}
	idxmap := mapping.IndexMappings()
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

	indices := cacheindex.NewIndices()
	indices.Add(maputils.Values(all)...)
	return d.factory.Apply(ctx, d, clusters, indices, logger)
}

func registerIndex[I cacheindex.Index](logger logging.Logger, i cacheindex.Definition, clusters types.Clusters, idxmap types.Mappings, mgr types.ControllerManager, local map[string]I) (cacheindex.Index, error) {
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
