package cluster

import (
	"context"
	"fmt"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils/cluster/config"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/options/activationopts"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
)

type ClusterFilter interface {
	GetUsedClusters(controllers types.ControllerNames) ClusterNames
}

func From(opts flagutils.OptionSetProvider) Definitions {
	return flagutils.GetFrom[Definitions](opts)
}

type Definitions = types.ClusterDefinitions

type definitions struct {
	internal.DefinitionsImpl[Definition, Definitions]
	scheme   *runtime.Scheme
	main     Definition
	clusters *clusters
}

var _ Definitions = (*definitions)(nil)

func NewDefinitions() Definitions {
	d := &definitions{
		main: Define(DEFAULT, "standard cluster", config.DefaultRules()),
	}
	d.DefinitionsImpl = internal.NewDefinitions[Definition, Definitions]("cluster", d)
	return d
}

func (d *definitions) WithScheme(scheme *runtime.Scheme) Definitions {
	d.scheme = scheme
	return d
}

func (d *definitions) GetScheme() *runtime.Scheme {
	return d.scheme
}

func (d *definitions) AddFlags(fs *pflag.FlagSet) {
	if d.Len() > 1 {
		// If we work with multiple clusters we enforce the usage og identity options
		d.main.RequireIdentity()
		for _, c := range d.Elements {
			c.RequireIdentity()
		}
	}
	d.main.AddFlags(fs)
	d.DefinitionsImpl.AddFlags(fs)
}

func (d *definitions) Validate(ctx context.Context, opts flagutils.OptionSet, v flagutils.ValidationSet) error {
	// don't establish validation dependencies, just access some metadata.
	filter := flagutils.GetFrom[ClusterFilter](opts)
	required := ClusterNames{}
	if filter != nil {
		copts := activationopts.From(opts)
		if copts != nil {
			required.AddAll(filter.GetUsedClusters(copts.GetActivation()))
		}
	}
	if len(required) == 0 {
		required.Add(d.GetNames()...)
	}

	intermediate := ClusterNames{}

	if d.clusters == nil && d.GetError() == nil {
		err := v.ValidateSet(ctx, opts, &d.DefinitionsImpl)
		if err != nil {
			return d.AddError(err, "validation")
		}
		d.clusters = newClusters()

		missing := true
		found := true
		for missing && found {
			missing = false
			found = false
			for n, def := range d.Elements {
				if d.clusters.Get(n) != nil || !(required.Contains(n) || intermediate.Contains(n)) {
					continue
				}

				acc, err := def.Create(d)
				if err != nil {
					return d.AddError(err, "cluster ", n)
				}
				if acc != nil {
					found = true
					d.clusters.Add(acc)
					continue
				}

				fb := def.GetFallback()
				if fb == "" {
					missing = true
					continue
				}
				eff := d.clusters.Get(fb)
				if eff != nil {
					if !def.AcceptFleet() && eff.AsFleet() != nil && required.Contains(def.GetName()) {
						return fmt.Errorf("fallback %q for cluster %q is fleet", fb, def.GetName())
					}
					d.clusters.Add(NewAlias(n, eff))
					found = true
				} else {
					intermediate.Add(fb)
					if fb == DEFAULT {
						err := v.Validate(ctx, opts, d.main)
						if err != nil {
							return d.AddError(err, "cluster ", d.main)
						}

						acc, err := d.main.Create(d)
						if err != nil {
							return d.AddError(err, "cluster ", d.main)
						}
						if acc != nil {
							d.clusters.Add(acc)
							found = true
							continue
						}
					}
					missing = true
				}
			}
		}
		if missing {
			for n, _ := range d.Elements {
				if d.clusters.Get(n) == nil {
					return d.AddError(fmt.Errorf("kubeconfig required"), "cluster ", n)
				}
			}
		}

		for n := range d.Elements {
			if d.clusters.Get(n) == nil {
				d.clusters.disabled.Add(n)
			}
		}
	}
	return d.GetError()
}

func (d *definitions) GetClusters() Clusters {
	return d.clusters
}

func ValidatedClusters(ctx context.Context, opts flagutils.OptionSet, v flagutils.ValidationSet) (Clusters, error) {
	defs, err := flagutils.ValidatedOptions[Definitions](ctx, opts, v)
	if err != nil {
		return nil, err
	}
	return defs.GetClusters(), nil
}
