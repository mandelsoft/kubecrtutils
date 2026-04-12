package ctrlmgmt

import (
	"context"
	"fmt"

	"github.com/mandelsoft/flagutils"
	ctrl "sigs.k8s.io/controller-runtime"
)

type runner struct {
	def Definition
}

func (r runner) Run(ctx context.Context, opts flagutils.OptionSet) error {
	mgr, err := r.def.GetControllerManager(ctx, opts)
	if err != nil {
		return err
	}
	mgr.GetLogger().Info("starting manager")
	return mgr.GetManager().Start(ctrl.SetupSignalHandler())
}

var _ flagutils.Runner = (*runner)(nil)

// Setup instantiates and runs a controller manager from its definition.
// The definition may either be passed as argument, or is taken from the option set.
func Setup(name string, op flagutils.OptionSetProvider, def Definition, args ...string) error {
	var opts flagutils.OptionSet
	if op != nil {
		opts = op.AsOptionSet()
	}
	if opts == nil {
		opts = &flagutils.DefaultOptionSet{}
	}
	found := From(opts)
	if found != nil {
		if def != nil && found != def {
			return fmt.Errorf("options already contain a controller manager definition")
		}
	} else {
		if def == nil {
			return fmt.Errorf("controller manager definition neither provided nor found in options")
		}
		found = def
		opts = flagutils.NewOptionSet(opts.AsOptionSet(), found)
	}

	return flagutils.ExecuteLifecycle(context.Background(), name, opts, runner{found}, args...)
}
