package cluster

import (
	"fmt"

	"github.com/mandelsoft/kubecrtutils/cluster/config"
	"github.com/mandelsoft/kubecrtutils/cluster/fleet"
	"github.com/mandelsoft/kubecrtutils/setup"
)

type fleetDef struct {
	baseDef[*fleetDef]
	typ      fleet.Type
	required bool
}

var _ Definition = (*fleetDef)(nil)

func DefineFleet(name, desc string, typ fleet.Type) *fleetDef {
	def := &fleetDef{typ: typ}
	def.baseDef = newBase[*fleetDef](def, name, desc)
	def.rules = typ.GetRules(def)
	return def
}

func (def *fleetDef) MustBeFleet() *fleetDef {
	def.required = true
	return def.self
}

////////////////////////////////////////////////////////////////////////////////

func (d *fleetDef) AcceptFleet() bool {
	return true
}

func (d *fleetDef) RequireFleet() bool {
	return d.required
}

func (d *fleetDef) GetFleetType() fleet.Type {
	return d.typ
}

func (d *fleetDef) Create(defs Definitions) (ClusterEquivalent, error) {
	ropts := &config.ConfigOptions{}
	cfg, err := d.GetConfig(ropts)
	if err != nil {
		return nil, fmt.Errorf("cluster %s: %w", d.name, err)
	}
	if cfg == nil {
		return nil, nil
	}
	c, err := d.typ.Create(defs, d, *cfg, setup.Log.WithName("fleet").WithName(d.GetName()))
	if err != nil || c != nil {
		return c, err
	}
	// fallback to regular cluster
	if d.RequireFleet() {
		return nil, fmt.Errorf("feet configuration for type %q missing", d.typ.GetType())
	}
	return d.asClusterDef().Create(defs)
}

func (d *fleetDef) asClusterDef() *clusterDef {
	c := &clusterDef{}
	c.baseDef = mapBase(d.baseDef, c)
	return c
}
