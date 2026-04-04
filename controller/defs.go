package controller

import (
	"context"
	"fmt"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/component"
	"github.com/mandelsoft/kubecrtutils/controller/constraints"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/options/activationopts"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/spf13/pflag"
)

func From(opts flagutils.OptionSetProvider) Definitions {
	return flagutils.GetFrom[Definitions](opts)
}

type filter struct {
	list ControllerNames
}

func (f *filter) Use(name string) bool {
	if f.list == nil {
		return true
	}
	return f.list.Has(name)
}

func (f *filter) Filter(opts flagutils.OptionSet) flagutils.OptionSet {
	if f.list == nil {
		return opts
	}
	filtered := flagutils.NewOptionSet()
	for o := range opts.Options {
		if d, ok := o.(Definition); ok {
			if f.Use((d.GetName())) {
				filtered.Add(o)
			}
		} else {
			filtered.Add(o)
		}
	}
	return filtered
}

type _definitions struct {
	internal.DefinitionsImpl[Definition, Definitions]
	groups      map[string][]string
	constraints constraints.Constraints
	filter      filter
}

var _ Definitions = (*_definitions)(nil)

func NewDefinitions() Definitions {
	d := &_definitions{constraints: constraints.New(), groups: make(map[string][]string)}
	d.DefinitionsImpl = internal.NewDefinitions[Definition, Definitions]("controller", d)
	return d
}

func (d *_definitions) Validate(ctx context.Context, opts flagutils.OptionSet, v flagutils.ValidationSet) error {
	// catch filter option, if present
	copts, err := flagutils.ValidatedOptions[*activationopts.Options](ctx, opts, v)
	if err != nil {
		return err
	}
	if copts != nil {
		d.filter.list = copts.GetActivation()
	}
	coopts, err := flagutils.ValidatedOptions[component.Definitions](ctx, opts, v)
	if err != nil {
		return err
	}
	// ToDo: component mapping
	for n, c := range d.Elements {
		for u := range c.GetComponents() {
			if coopts == nil || coopts.Get(u) == nil {
				return fmt.Errorf("conroller %q used unknown component %q", n, u)
			}
		}
	}

	return v.ValidateSet(ctx, opts, d.AsOptionSet()) // forward validation
}

func (d *_definitions) AddRule(constraints ...types.Constraint) Definitions {
	d.constraints.Add(constraints...)
	return d
}

func (d *_definitions) Add(elems ...Definition) Definitions {
	for _, e := range elems {
		for n := range e.GetGroups() {
			d.groups[n] = append(d.groups[n], e.GetName())
		}
	}
	return d.DefinitionsImpl.Add(elems...)
}

func (d *_definitions) GetControllerSet() activationopts.ControllerSet {
	return d
}

func (d *_definitions) GetGroups() map[string][]string {
	return d.groups
}

func (d *_definitions) AddFlags(fs *pflag.FlagSet) {
	d.DefinitionsImpl.AddFlags(fs)
}

func (d *_definitions) GetUsedClusters(ctx constraints.Context) cluster.ClusterNames {
	names := set.New[string]()
	for n, c := range d.Elements {
		if ctx.Has(n) {
			names.AddAll(c.GetRequiredClusters(nil))
		}
	}
	return names
}

func (d *_definitions) GetUsedComponents(ctx constraints.Context) component.ComponentNames {
	names := set.New[string]()
	for n, c := range d.Elements {
		if ctx.Has(n) {
			names.AddAll(c.GetRequiredComponents(nil))
		}
	}
	return names
}

func (d *_definitions) CreateIndices(ctx context.Context, mgr types.ControllerManager) error {
	mgr.GetLogger().Info("configure controller defined indices...")
	if d.GetError() != nil {
		return d.GetError()
	}

	list := d.filter.list
	if list == nil {
		list = set.New[string](d.GetNames()...)
	}
	cctx := constraints.NewContext(d.GetControllerSet()).WithSelectedSet(list)
	_, err := d.constraints.Match(cctx)
	if err != nil {
		return err
	}

	for _, c := range d.Elements {
		if !d.filter.Use(c.GetName()) {
			continue
		}
		cset := c.GetActivationConstraints()
		if cset.Len() >= 0 {
			_, err := cset.Match(cctx)
			if err != nil {
				return err
			}
		}
	}

	for _, c := range d.Elements {
		if !d.filter.Use(c.GetName()) {
			continue // should we create indices if they are required elsewhere
		}
		err := c.CreateIndices(ctx, nil, mgr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *_definitions) Apply(ctx context.Context, mgr types.ControllerManager) (Controllers, error) {
	mgr.GetLogger().Info("configure controllers...")
	if d.GetError() != nil {
		return nil, d.GetError()
	}

	controllers := NewControllers()
	// Step 1: create controllers and their environment like indices
	for n, i := range d.Elements {
		if !d.filter.Use(n) {
			continue
		}
		c, err := i.Apply(ctx, nil, mgr)
		if err != nil {
			return nil, fmt.Errorf("controller %q: %w", n, err)
		}
		err = controllers.Add(c)
		if err != nil {
			return nil, fmt.Errorf("controller %q: %w", n, err)
		}
	}
	// Step 2: complete the controller by creating their reconciler
	// (by factory) and finally configure the controller runtime manager.
	for n, i := range controllers.Elements {
		if !d.filter.Use(n) {
			continue
		}
		err := i.Complete(ctx)
		if err != nil {
			return nil, fmt.Errorf("controller %q: %w", n, err)
		}
	}
	return controllers, nil
}

func (d *_definitions) AsOptionSet() flagutils.OptionSet {
	if d.filter.list == nil {
		return d.DefinitionsImpl.AsOptionSet()
	}
	return d.filter.Filter(d.DefinitionsImpl.AsOptionSet())
}
