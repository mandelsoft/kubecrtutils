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
	"github.com/mandelsoft/kubecrtutils/utils"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type ComponentFilter interface {
	GetUsedComponents(constraints.Context) ComponentNames
}

func From(opts flagutils.OptionSetProvider) Definitions {
	return flagutils.GetFrom[Definitions](opts)
}

type Definitions interface {
	internal.Definitions[Definition, Definitions]
	cluster.ClusterFilter

	flagutils.Validatable

	CreateIndices(ctx context.Context, mgr types.ControllerManager) error
	Apply(ctx context.Context, mgr types.ControllerManager) error
}

type _definitions struct {
	internal.DefinitionsImpl[Definition, Definitions]

	cctx     constraints.Context
	required ComponentNames
}

func NewDefinitions() Definitions {
	d := &_definitions{}
	d.DefinitionsImpl = internal.NewDefinitions[Definition, Definitions]("component", d)
	return d
}

func (d *_definitions) isUsed(c Definition) (bool, error) {
	cset := c.GetActivationConstraints()
	if cset != nil {
		ok, err := c.GetActivationConstraints().Match(d.cctx)
		if err != nil {
			return false, err
		}
		if ok == constraints.Yes {
			return true, nil
		}
		if ok == constraints.No {
			return false, nil
		}
	}
	return d.required.Contains(c.GetName()), nil
}

func (d *_definitions) Validate(ctx context.Context, opts flagutils.OptionSet, v flagutils.ValidationSet) error {
	_, err := flagutils.ValidatedOptions[types.ControllerDefinition](ctx, opts, v)
	if err != nil {
		return err
	}

	d.required = utils.GetUsed[ComponentFilter, ComponentNames](opts)

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

func (d *_definitions) GetUsedClusters(ctx constraints.Context) cluster.ClusterNames {
	names := set.New[string]()
	for _, c := range d.Elements {
		if ok, _ := d.isUsed(c); ok {
			names.AddAll(c.GetRequiredClusters(nil))
		}
	}
	return names
}

func (d *_definitions) GetUsedComponents(ctx constraints.Context) ComponentNames {
	names := set.New[string]()
	mod := true
	for mod {
		mod = false
		for _, c := range d.Elements {
			if ok, _ := d.isUsed(c); ok {
				l := len(names)
				names.AddAll(c.GetRequiredComponents(nil))
				mod = mod || l != len(names)
			}
		}
	}
	return names
}

func (d *_definitions) CreateIndices(ctx context.Context, mgr types.ControllerManager) error {
	for _, c := range d.Elements {
		if ok, err := d.isUsed(c); !ok || err != nil {
			if err != nil {
				return err
			}
			continue
		}
		err := c.CreateIndices(ctx, nil, mgr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *_definitions) Apply(ctx context.Context, mgr types.ControllerManager) error {
	mgr.GetLogger().Info("configure components...")

	for n, c := range d.Elements {
		if ok, err := d.isUsed(c); !ok || err != nil {
			continue
		}

		comp, err := c.Apply(ctx, nil, mgr)
		if err != nil {
			return fmt.Errorf("component %q: %w", n, err)
		}
		err = mgr.GetComponents().Add(comp)
		if err != nil {
			return fmt.Errorf("component %q: %w", n, err)
		}
		if r, ok := comp.(manager.Runnable); ok {
			mgr.GetLogger().Info("  register as runnable")
			mgr.GetManager().GetLocalManager().Add(r)
		}
	}
	return nil
}
