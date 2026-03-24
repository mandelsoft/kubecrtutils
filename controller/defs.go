package controller

import (
	"context"
	"fmt"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/controller/rules"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/options/activationopts"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/spf13/pflag"
)

func From(opts flagutils.OptionSetProvider) Definitions {
	return flagutils.GetFrom[Definitions](opts)
}

type Definitions interface {
	internal.Definitions[Definition, Definitions]
	AddRule(rules ...rules.Rule) Definitions

	flagutils.Validatable

	activationopts.ControllerSource
	cluster.ClusterFilter

	Apply(ctx context.Context, manager types.ControllerManager) (Controllers, error)
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

type _definitions struct {
	internal.DefinitionsImpl[Definition, Definitions]
	groups map[string][]string
	rules  rules.Rules
	filter filter
}

var _ Definitions = (*_definitions)(nil)

func NewDefinitions() Definitions {
	d := &_definitions{rules: rules.New()}
	d.DefinitionsImpl = internal.NewDefinitions[Definition, Definitions]("index", d)
	return d
}

func (d *_definitions) Validate(ctx context.Context, opts flagutils.OptionSet, v flagutils.ValidationSet) error {
	// catch filter option, if present
	copts, err := flagutils.ValidatedOptions[*activationopts.Options](ctx, opts, v)
	if err != nil {
		return err
	}
	if copts == nil {
		d.filter.list = copts.GetActivation()
	}
	return v.ValidateSet(ctx, opts, d.AsOptionSet()) // forward validation
}

func (d *_definitions) AddRule(rules ...rules.Rule) Definitions {
	d.rules.Add(rules...)
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

func (d *_definitions) GetUsedClusters() ClusterNames {
	names := set.New[string]()
	for _, c := range d.Elements {
		names.AddAll(c.GetRequiredClusters(nil))
	}
	return names
}

func (d *_definitions) Apply(ctx context.Context, mgr types.ControllerManager) (Controllers, error) {
	mgr.GetLogger().Info("configure controller defined indices...")
	if d.GetError() != nil {
		return nil, d.GetError()
	}
	if d.rules.Len() > 0 {
		list := d.filter.list
		if list == nil {
			list = set.New[string](d.GetNames()...)
		}
		err := d.rules.Match(rules.NewContext(d.GetControllerSet()), list)
		if err != nil {
			return nil, err
		}
	}
	for _, c := range d.Elements {
		if !d.filter.Use(c.GetName()) {
			continue // should we create indices if they are required elsewhere
		}
		err := c.CreateIndices(ctx, nil, mgr)
		if err != nil {
			return nil, err
		}
	}

	controllers := NewControllers()
	mgr.GetLogger().Info("configure controllers...")
	// Step 1: create controllers and their environment like indices
	for n, i := range d.Elements {
		if !d.filter.Use(n) {
			continue
		}
		c, err := i.CreateController(ctx, nil, mgr)
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
