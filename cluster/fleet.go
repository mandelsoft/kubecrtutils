package cluster

import (
	"fmt"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/kubecrtutils/cluster/config"
	"github.com/mandelsoft/kubecrtutils/fleet"
)

type fleetdef struct {
	definition
	typ fleet.Type
}

var _ Definition = (*fleetdef)(nil)

func DefineFleet(name, desc string, typ fleet.Type, required ...bool) Definition {
	def := &definition{name: name, desc: desc, fallback: DEFAULT, fleet: general.Optional(required...)}
	def.rules = typ.GetRules(def)
	return def
}

func (d *fleetdef) GetFleetType() fleet.Type {
	return d.typ
}

func (d *fleetdef) Create(defs Definitions) (ClusterEquivalent, error) {
	ropts := &config.ConfigOptions{}
	cfg, err := d.GetConfig(ropts)
	if err != nil {
		return nil, fmt.Errorf("cluster %s: %w", d.name, err)
	}
	if cfg == nil {
		return nil, nil
	}
	c, err := d.typ.Create(defs, d, *cfg)
	if err != nil || c != nil {
		return c, err
	}
	// fallback to regular cluster
	if d.RequireFleet() {
		return nil, fmt.Errorf("feet configuration for type %q missing", d.typ.GetType())
	}
	return d.definition.Create(defs)
}
