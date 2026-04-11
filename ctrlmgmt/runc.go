package ctrlmgmt

import (
	"context"
	"fmt"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/pflag"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Setup(name string, opts flagutils.OptionSet, def Definition, args ...string) error {
	fs := pflag.NewFlagSet(name, pflag.ContinueOnError)
	ctx := context.Background()

	found := From(opts)
	if found != nil {
		if def != nil && found != def {
			return fmt.Errorf("options already contain a controller manager definition")
		}
	} else {
		if def == nil {
			return fmt.Errorf("controller manager definition neither provided nor found in options")
		}
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
	err = mgr.GetManager().Start(ctrl.SetupSignalHandler())

	return errors.Join(err, flagutils.Finalize(ctx, opts, nil))
}
