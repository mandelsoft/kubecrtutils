package cluster

import (
	"fmt"

	"github.com/mandelsoft/kubecrtutils/cluster/cluster"
	"github.com/mandelsoft/kubecrtutils/cluster/config"
	sigcluster "sigs.k8s.io/controller-runtime/pkg/cluster"
)

type clusterDef struct {
	baseDef[*clusterDef]
}

var _ Definition = (*clusterDef)(nil)

func Define(name string, desc string, rule ...config.Rule) *clusterDef {
	d := &clusterDef{}
	d.baseDef = newBase[*clusterDef](d, name, desc, rule...)
	return d
}

func (d *clusterDef) AcceptFleet() bool {
	return false
}

func (d *clusterDef) Create(defs Definitions) (ClusterEquivalent, error) {
	ropts := &config.ConfigOptions{}
	cfg, err := d.GetConfig(ropts)
	if err != nil {
		return nil, fmt.Errorf("cluster %s: %w", d.name, err)
	}
	if cfg == nil {
		return nil, nil
	}
	c, err := cluster.NewCluster(d.name, cfg, func(opts *sigcluster.Options) {
		if d.scheme != nil {
			opts.Scheme = d.scheme
		} else {
			opts.Scheme = defs.GetScheme()
		}
	})
	return c, nil
}
