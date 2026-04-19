package controller

import (
	"context"

	"github.com/mandelsoft/flagutils"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

// --- begin config interface ---

type ControllerOptions = controller.TypedOptions[mcreconcile.Request]

// ConfigurationProvider is used to modify the controller options used to
// create a new controller.
// Such objects can be set at the Options object to preprocess the configuration
// after the option parsing. Or this interface can be implemented by other
// Options types to incorporate their settings into the configuration.
type ConfigurationProvider interface {
	ConfigureController(ctx context.Context, config *ControllerOptions, name string, opts flagutils.OptionSet) error
}

// --- end config interface ---
