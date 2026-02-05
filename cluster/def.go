package cluster

import (
	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils/cluster/config"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
)

const DEFAULT = "default"

type DefinitionProvider = types.ClusterDefinitionProvider

type Definition = types.ClusterDefinition

type baseAttr struct {
	name     string
	fallback string
	rules    config.Rules
	desc     string
	scheme   *runtime.Scheme
}

type baseDef[D Definition] struct {
	self D
	baseAttr
}

func newBase[D Definition](self D, name string, desc string, rule ...config.Rule) baseDef[D] {
	if len(rule) == 0 {
		rule = []config.Rule{config.DedicatedConfigRules(name, desc)}
	}
	return baseDef[D]{self: self, baseAttr: baseAttr{name: name, desc: desc, fallback: DEFAULT, rules: config.NewRules(rule...)}}
}

func (d *baseDef[D]) WithFallback(fallback string) D {
	d.fallback = fallback
	return d.self
}

func (d *baseDef[D]) WithScheme(scheme *runtime.Scheme) D {
	d.scheme = scheme
	return d.self
}

////////////////////////////////////////////////////////////////////////////////

func (d *baseDef[D]) GetDefinition() Definition {
	return d.self
}

func (d *baseDef[F]) RequireIdentity() {
	d.rules.RequireIdentity()
}

func (d *baseDef[D]) GetName() string {
	return d.name
}

func (d *baseDef[D]) GetFallback() string {
	return d.fallback
}

func (d *baseDef[D]) GetDescription() string {
	return d.desc
}

func (d *baseDef[D]) GetScheme() *runtime.Scheme {
	return d.scheme
}

func (d *baseDef[D]) GetConfig(o *config.ConfigOptions) (*config.Config, error) {
	return d.rules.GetConfig(o)
}

func (d *baseDef[D]) AddFlags(fs *pflag.FlagSet) {
	d.rules.AddFlags(fs)
}

func (d *baseDef[D]) AsOptionSet() flagutils.OptionSet {
	return d.rules.AsOptionSet()
}

func mapBase[I, O Definition](in baseDef[I], o O) baseDef[O] {
	return baseDef[O]{
		self:     o,
		baseAttr: in.baseAttr,
	}
}
