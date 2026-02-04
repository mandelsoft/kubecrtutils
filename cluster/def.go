package cluster

import (
	"fmt"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils/cluster/config"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

const DEFAULT = "default"

type DefinitionProvider = types.ClusterDefinitionProvider

type Definition = types.ClusterDefinition

type definition struct {
	name     string
	fallback string
	rules    config.Rules
	desc     string
	scheme   *runtime.Scheme
	fleet    bool
}

var _ Definition = (*definition)(nil)

func Define(name string, desc string, rule ...config.Rule) Definition {
	if len(rule) == 0 {
		rule = []config.Rule{config.DedicatedConfigRules(name, desc)}
	}
	return &definition{name: name, desc: desc, fallback: DEFAULT, rules: config.NewRules(rule...)}
}

func (d *definition) RequireFleet() bool {
	return d.fleet
}

func (d *definition) WithFallback(fallback string) Definition {
	d.fallback = fallback
	return d
}

func (d *definition) WithScheme(scheme *runtime.Scheme) Definition {
	d.scheme = scheme
	return d
}

func (d *definition) GetDefinition() Definition {
	return d
}

func (d *definition) RequireIdentity() {
	d.rules.RequireIdentity()
}

func (d *definition) GetName() string {
	return d.name
}

func (d *definition) GetFallback() string {
	return d.fallback
}

func (d *definition) GetDescription() string {
	return d.desc
}

func (d *definition) GetScheme() *runtime.Scheme {
	return d.scheme
}

func (d *definition) GetConfig(o *config.ConfigOptions) (*config.Config, error) {
	return d.rules.GetConfig(o)
}

func (d *definition) AddFlags(fs *pflag.FlagSet) {
	d.rules.AddFlags(fs)
}

func (d *definition) AsOptionSet() flagutils.OptionSet {
	return d.rules.AsOptionSet()
}

func (d *definition) Create(defs Definitions) (ClusterEquivalent, error) {
	ropts := &config.ConfigOptions{}
	cfg, err := d.GetConfig(ropts)
	if err != nil {
		return nil, fmt.Errorf("cluster %s: %w", d.name, err)
	}
	if cfg == nil {
		return nil, nil
	}
	c, err := NewCluster(d.name, cfg, func(opts *cluster.Options) {
		if d.scheme != nil {
			opts.Scheme = d.scheme
		} else {
			opts.Scheme = defs.GetScheme()
		}
	})
	return c, nil
}
