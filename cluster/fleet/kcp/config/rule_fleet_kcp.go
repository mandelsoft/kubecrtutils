package config

import (
	"fmt"

	"github.com/mandelsoft/flagutils"
	config2 "github.com/mandelsoft/kubecrtutils/cluster/config"
	"github.com/spf13/pflag"
)

type KCPFleetOption struct {
	name string

	endpointslice string
}

var (
	_ flagutils.Options      = (*KCPFleetOption)(nil)
	_ config2.Rule           = (*KCPFleetOption)(nil)
	_ config2.Personalizable = (*KCPFleetOption)(nil)
)

func NewKCPFleetOption(name string) *KCPFleetOption {
	if name == "" {
		name = "kubeconfig"
	}
	return &KCPFleetOption{
		name: name,
	}
}

func (r *KCPFleetOption) PersonalizedWith(o *config2.Personalization) config2.Rule {
	name := o.Name
	if len(name) == 0 {
		name = r.name
	} else {
		name = name + "-" + r.name
	}
	return &KCPFleetOption{
		name: name,
	}
}

func (r *KCPFleetOption) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&r.endpointslice, r.name+"-endpointslice", "", "", "endpointslice used together with "+r.name+"for APIExport")
}

func (r *KCPFleetOption) GetConfig(opts *config2.ConfigOptions) (*config2.Config, error) {
	if r.endpointslice != "" {
		if opts.SubConfig != nil {
			return nil, fmt.Errorf("cannot specify both endpointslices with subconfig of type '%s'", opts.SubConfig.GetType())
		}
		opts.SubConfig = &KCPFleetConfig{
			EndpointSlice: r.endpointslice,
		}
	}
	return nil, nil
}
