package hostedzone

import (
	"context"
	"fmt"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/options/manageropts"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Options struct {
	Class string
}

func From(opts flagutils.OptionSetProvider) *Options {
	return flagutils.GetFrom[*Options](opts)
}

var (
	_ flagutils.Options                 = (*Options)(nil)
	_ flagutils.Validatable             = (*Options)(nil)
	_ manageropts.ConfigurationProvider = (*Options)(nil)
)

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) Validate(ctx context.Context, opts flagutils.OptionSet, v flagutils.ValidationSet) error {
	var err error

	clusters, err := cluster.ValidatedClusters(ctx, opts, v)
	if err != nil {
		return err
	}
	if clusters.Get("dataplane") == nil {
		return fmt.Errorf("dataplane cluster is required")
	}
	if clusters.Get("runtime") == nil {
		return fmt.Errorf("dataplane cluster is required")
	}

	return nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Class, "class", "", "", "name of the controller class to handle")
}

func (o *Options) Configure(ctx context.Context, cfg *manager.Options, opts flagutils.OptionSet) error {

	return nil
}
