package component

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/controller/constraints"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/options/activationopts"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/mandelsoft/kubecrtutils/utils"
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

	types.IndexProvider
	types.Applyable
}

type _definitions struct {
	internal.DefinitionsImpl[Definition, Definitions]

	cctx     constraints.Context
	required ComponentNames
	order    []string
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

	order, cycle := utils.TopoSort(d.Elements, d.deps)
	if cycle != nil {
		return fmt.Errorf("dependency cyle for components: %s", strings.Join(cycle, "->"))
	}
	d.order = order
	return v.ValidateSet(ctx, opts, d.AsOptionSet()) // forward validation
}

func (d *_definitions) deps(def Definition) iter.Seq2[string, Definition] {
	return func(yield func(string, Definition) bool) {
		for k := range def.GetComponents() {
			if !yield(k, d.Get(k)) {
				return
			}
		}
	}
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

func (d *_definitions) CreateIndices(ctx context.Context, mappings mapping.ControllerMappings, mgr types.ControllerManager) error {
	for _, c := range d.Elements {
		if ok, err := d.isUsed(c); !ok || err != nil {
			if err != nil {
				return err
			}
			continue
		}
		err := c.CreateIndices(ctx, mappings, mgr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *_definitions) Apply(ctx context.Context, mappings mapping.ControllerMappings, mgr types.ControllerManager) error {
	mgr.GetLogger().Info("configure components...")

	for _, n := range d.order {
		c := d.Get(n)
		if ok, err := d.isUsed(c); !ok || err != nil {
			continue
		}

		err := c.Apply(ctx, mappings, mgr)
		if err != nil {
			return fmt.Errorf("component %q: %w", n, err)
		}
	}
	return nil
}
