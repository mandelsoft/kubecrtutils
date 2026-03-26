package ctrlmgmt

import (
	"context"
	"fmt"

	"github.com/mandelsoft/flagutils"
	"github.com/spf13/pflag"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Setup(name string, opts flagutils.OptionSet, def Definition, args ...string) error {
	fs := pflag.NewFlagSet(name, pflag.ContinueOnError)
	ctx := context.Background()

	found := From(opts)
	if found != nil {
		if found != def {
			return fmt.Errorf("options already contain a definition")
		}
	} else {
		opts = flagutils.NewOptionSet(opts, def)
	}
	if err := flagutils.Prepare(ctx, opts, nil); err != nil {
		return err
	}
	opts.AddFlags(fs)

	err := fs.Parse(args)
	if err != nil {
		return err
	}
	err = flagutils.Validate(ctx, opts, nil)
	if err != nil {
		return err
	}
	mgr, err := def.GetControllerManager(ctx, opts)
	if err != nil {
		return err
	}
	mgr.GetLogger().Info("starting manager")
	return mgr.GetManager().Start(ctrl.SetupSignalHandler())
}
