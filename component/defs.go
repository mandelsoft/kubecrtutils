package component

import (
	"context"
	"fmt"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/controller/constraints"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/options/activationopts"
	"github.com/mandelsoft/kubecrtutils/types"
)

func From(opts flagutils.OptionSetProvider) Definitions {
	return flagutils.GetFrom[Definitions](opts)
}

type Definitions interface {
	internal.Definitions[Definition, Definitions]
	cluster.ClusterFilter

	flagutils.Validatable

	CreateIndices(ctx context.Context, mgr types.ControllerManager) error
	Apply(ctx context.Context, mgr types.ControllerManager) (Components, error)
}

type _definitions struct {
	internal.DefinitionsImpl[Definition, Definitions]

	cctx *constraints.Context
}

func NewDefinitions() Definitions {
	d := &_definitions{}
	d.DefinitionsImpl = internal.NewDefinitions[Definition, Definitions]("index", d)
	return d
}

func (d *_definitions) Validate(ctx context.Context, opts flagutils.OptionSet, v flagutils.ValidationSet) error {
	// catch filter option, if present
	copts, err := flagutils.ValidatedOptions[*activationopts.Options](ctx, opts, v)
	if err != nil {
		return err
	}
	if copts != nil {
		d.cctx = copts.GetContraintContext()
	}
	return v.ValidateSet(ctx, opts, d.AsOptionSet()) // forward validation
}

func (d *_definitions) GetUsedClusters(ctx *constraints.Context) cluster.ClusterNames {
	names := set.New[string]()
	for n, c := range d.Elements {
		cond, err := c.GetActivationConstraints().Match(ctx)
		d.AddError(err, "constraints for ", n)
		if cond == constraints.Yes {
			names.AddAll(c.GetRequiredClusters(nil))
		}
	}
	return names
}

func (d *_definitions) CreateIndices(ctx context.Context, mgr types.ControllerManager) error {
	for _, c := range d.Elements {
		cset := c.GetActivationConstraints()
		if cset != nil {
			ok, err := c.GetActivationConstraints().Match(d.cctx)
			if err != nil {
				return err
			}
			if ok != constraints.Yes {
				continue
			}
		}
		err := c.CreateIndices(ctx, nil, mgr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *_definitions) Apply(ctx context.Context, mgr types.ControllerManager) (Components, error) {
	mgr.GetLogger().Info("configure components...")

	comps := NewComponents()
	for n, c := range d.Elements {
		cset := c.GetActivationConstraints()
		if cset != nil {
			ok, err := c.GetActivationConstraints().Match(d.cctx)
			if err != nil {
				return nil, err
			}
			if ok != constraints.Yes {
				continue
			}
		}

		comp, err := c.CreateComponent(ctx, nil, mgr)
		if err != nil {
			return nil, fmt.Errorf("component %q: %w", n, err)
		}
		err = comps.Add(comp)
		if err != nil {
			return nil, fmt.Errorf("component %q: %w", n, err)
		}
	}
	return comps, nil
}
