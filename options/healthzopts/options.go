package healthzopts

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils/options/manageropts"
	"github.com/spf13/pflag"
	ctrl "sigs.k8s.io/multicluster-runtime"
)

type Options struct {
	HealthProbeBindAddress string
}

var _ manageropts.ConfigurationProvider = (*Options)(nil)

func From(opts flagutils.OptionSetProvider) *Options {
	return flagutils.GetFrom[*Options](opts)
}

var (
	_ flagutils.Options     = (*Options)(nil)
	_ flagutils.Validatable = (*Options)(nil)
)

func New() *Options {
	return &Options{}
}

func (o *Options) Validate(ctx context.Context, opts flagutils.OptionSet, v flagutils.ValidationSet) error {
	fields := strings.Split(o.HealthProbeBindAddress, ":")

	if len(fields) != 2 {
		return fmt.Errorf("health probe bind address must contain ':<port>'")
	}
	_, err := strconv.Atoi(fields[1])
	if err != nil {
		return fmt.Errorf("port must be integer: %w", err)
	}
	return nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.HealthProbeBindAddress, "health-probe-bind-address", ":8081", "The address the healthz endpoint binds to.")
}

func (o *Options) Configure(ctx context.Context, config *ctrl.Options, opts flagutils.OptionSet) error {
	config.HealthProbeBindAddress = o.HealthProbeBindAddress
	return nil
}
