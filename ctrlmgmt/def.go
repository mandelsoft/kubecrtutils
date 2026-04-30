package ctrlmgmt

import (
	"errors"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/component"
	"github.com/mandelsoft/kubecrtutils/controller"
	"github.com/mandelsoft/kubecrtutils/controller/constraints"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/options/manageropts"
	"github.com/mandelsoft/kubecrtutils/owner"
	"github.com/spf13/pflag"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/runtime"
)

func From(opts flagutils.OptionSetProvider) Definition {
	return flagutils.GetFrom[Definition](opts)
}

type Definition interface {
	internal.Named
	flagutils.Options
	flagutils.OptionSetProvider

	GetController(name string) controller.Definition
	GetComponent(name string) component.Definition
	GetIndex(name string) cacheindex.Definition

	GetOwnerHandlerProvider() owner.HandlerProvider

	GetError() error
	GetControllerManager(ctx context.Context, opts flagutils.OptionSetProvider) (ControllerManager, error)
}

// --- begin definition ---

type CompositionInterface interface {
	Definition

	WithOwnerHandler(provider owner.HandlerProvider) CompositionInterface
	WithScheme(scheme *runtime.Scheme) CompositionInterface
	AddCluster(def ...cluster.Definition) CompositionInterface
	AddComponent(def ...component.Definition) CompositionInterface
	AddController(def ...controller.Definition) CompositionInterface
	AddControllerRule(rules ...constraints.Constraint) CompositionInterface
	AddIndex(def ...cacheindex.Definition) CompositionInterface
}

// --- end definition ---

type definition struct {
	internal.Element
	options     flagutils.DefaultOptionSet
	clusters    cluster.Definitions
	indices     cacheindex.Definitions
	components  component.Definitions
	controllers controller.Definitions
	owner       owner.HandlerProvider
}

var _ CompositionInterface = (*definition)(nil)

func Define(name string, main ...string) CompositionInterface {
	d := &definition{
		Element:     internal.NewElement(name),
		clusters:    cluster.NewDefinitions(),
		indices:     cacheindex.NewDefinitions(),
		components:  component.NewDefinitions(),
		controllers: controller.NewDefinitions(),
		owner:       owner.DefaultProvider,
	}
	d.options.Add(d.clusters, d.indices, d.controllers, d.components, manageropts.New(main, name))
	return d
}

func (d *definition) WithScheme(scheme *runtime.Scheme) CompositionInterface {
	d.clusters.WithScheme(scheme)
	return d
}

func (d *definition) WithOwnerHandler(h owner.HandlerProvider) CompositionInterface {
	d.owner = h
	return d
}

func (d *definition) AddCluster(def ...cluster.Definition) CompositionInterface {
	d.clusters.Add(def...)
	return d
}

func (d *definition) AddComponent(def ...component.Definition) CompositionInterface {
	d.components.Add(def...)
	return d
}

func (d *definition) AddController(def ...controller.Definition) CompositionInterface {
	d.controllers.Add(def...)
	return d
}

func (d *definition) AddControllerRule(rules ...constraints.Constraint) CompositionInterface {
	d.controllers.AddRule(rules...)
	return d
}

func (d *definition) AddIndex(def ...cacheindex.Definition) CompositionInterface {
	d.indices.Add(def...)
	return d
}

func (d *definition) AddFlags(fs *pflag.FlagSet) {
	for o := range d.options.Options {
		o.AddFlags(fs)
	}
}

func (d *definition) AsOptionSet() flagutils.OptionSet {
	return &d.options
}

func (d *definition) GetController(name string) controller.Definition {
	return d.controllers.Get(name)
}

func (d *definition) GetComponent(name string) component.Definition {
	return d.components.Get(name)
}

func (d *definition) GetIndex(name string) cacheindex.Definition {
	return d.indices.Get(name)
}

func (d *definition) GetOwnerHandlerProvider() owner.HandlerProvider {
	return d.owner
}

func (d *definition) GetError() error {
	return errors.Join(d.clusters.GetError(), d.indices.GetError(), d.controllers.GetError())
}

func (d *definition) GetControllerManager(ctx context.Context, opts flagutils.OptionSetProvider) (ControllerManager, error) {
	return NewControllerManagerByOpts(ctx, opts.AsOptionSet())
}
