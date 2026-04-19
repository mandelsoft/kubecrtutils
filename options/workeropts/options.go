package workeropts

import (
	"context"
	"fmt"
	"slices"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils/controller"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/spf13/pflag"
)

func From(set flagutils.OptionSet) *Options {
	return flagutils.GetFrom[*Options](set)
}

type Options struct {
	concurreny map[string]int
	set        controller.ControllerRefrences
}

var (
	_ flagutils.Options                = (*Options)(nil)
	_ flagutils.Validatable            = (*Options)(nil)
	_ controller.ConfigurationProvider = (*Options)(nil)
)

// New provides options usable to activate or deactivate controllers
func New() *Options {
	return &Options{}
}

func (o *Options) Validate(ctx context.Context, opts flagutils.OptionSet, v flagutils.ValidationSet) error {
	cdefs := flagutils.GetFrom[types.ControllerSource](opts)
	if cdefs == nil || len(cdefs.GetControllerSet().GetNames()) == 0 {
		if len(o.concurreny) != 0 {
			return fmt.Errorf("no controllers found to set workers for")
		}
	}

	set := cdefs.GetControllerSet()
	names := set.GetNames()
	groups := set.GetGroups()
	o.set = controller.CompleteSet(set)
	var invalid []string

	for name := range o.concurreny {
		if slices.Contains(names, name) {
			continue
		}
		if _, ok := groups[name]; ok {
			continue
		}
		invalid = append(invalid, name)
	}
	if len(invalid) > 0 {
		return fmt.Errorf("option workers: invalid controller names or groups: %v", invalid)
	}
	return nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringToIntVarP(&o.concurreny, "workers", "", nil, "workers for controller")
}

////////////////////////////////////////////////////////////////////////////////

func (o *Options) ConfigureController(ctx context.Context, config *controller.ControllerOptions, name string, opts flagutils.OptionSet) error {
	for n, i := range o.concurreny {
		s := o.set[n]
		if s != nil && s.Contains(name) {
			if i > 0 {
				config.Logger.Info("  setting worker count for {{controller}} to {{value}}",
					"controller", name, "value", i)
				config.MaxConcurrentReconciles = i
			}
		}
	}
	return nil
}
