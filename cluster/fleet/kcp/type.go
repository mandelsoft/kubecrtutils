package kcp

import (
	"fmt"

	"github.com/kcp-dev/multicluster-provider/apiexport"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/cluster/config"
	"github.com/mandelsoft/kubecrtutils/cluster/fleet"
	config3 "github.com/mandelsoft/kubecrtutils/cluster/fleet/kcp/config"
	"github.com/mandelsoft/logging"
)

type _type struct {
}

var _ fleet.Type = (*_type)(nil)

func Type() fleet.Type {
	return &_type{}
}

func (d *_type) GetType() string {
	return config3.SUBTYPE_KCPFLEET
}

func (d *_type) GetRules(def cluster.Definition) config.Rules {
	if def.GetName() == cluster.DEFAULT {
		return config.NewRules(config3.NewKCPFleetOption(""), config.DefaultRules())
	}
	return config.DedicatedConfigRules(def.GetName(), def.GetDescription(), config3.NewKCPFleetOption(""))
}

func (d *_type) Create(defs cluster.Definitions, def cluster.Definition, config config.Config, log logging.Logger) (fleet.Fleet, error) {
	if config.SubConfig == nil {
		return nil, nil
	}
	cfg, ok := config.SubConfig.(*config3.KCPFleetConfig)
	if !ok {
		return nil, fmt.Errorf("unexpected sub config of type %q", config.SubConfig.GetType())
	}
	return New(d, def.GetName(), general.OptionalNonZeroDefaulted(def.GetName(), config.GetId()), config.RestConfig, cfg.EndpointSlice, apiexport.Options{
		Scheme: general.OptionalDefaulted(defs.GetScheme(), def.GetScheme()),
		Log:    generics.PointerTo(log.V(0)),
	})
}
