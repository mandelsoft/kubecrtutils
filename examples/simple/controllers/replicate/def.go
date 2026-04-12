package replicate

import (
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/controller"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconciler/logic"
	"github.com/mandelsoft/kubecrtutils/examples/simple/controllers"
	"github.com/mandelsoft/logging"
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
)

// --- begin definition ---

type Resource = v1.ConfigMap

func Controller() controller.Definition {
	return controller.Define[*Resource](
		"replicate",
		controllers.SOURCE,
		logic.New[*controllers.Options, *controllers.Settings, *Resource](&ReconcilationLogic{}),
	).
		UseCluster(controllers.TARGET).
		AddTrigger(controller.OwnerTrigger[*Resource]().OnCluster(controllers.TARGET)).
		AddIndexByFactory("test", indexerFactory).
		AddForeignIndex(cacheindex.DefineByFactory[*Resource]("test", controllers.TARGET, indexerFactory))
}

// --- end definition ---

func indexerFactory(ctx context.Context, logger logging.Logger, clusters cluster.Clusters) (cacheindex.TypedIndexerFunc[*Resource], error) {
	opts := cacheindex.OptionsFromContext(ctx).(*controllers.Options)
	return func(obj *Resource) []string {
		annos := obj.GetAnnotations()
		if annos == nil {
			return nil
		}
		c, ok := annos[opts.Annotation]
		if !ok {
			return nil
		}
		return []string{c}
	}, nil
}
