package replicate

import (
	"github.com/mandelsoft/kubecrtutils/controller"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconciler/support"
	"github.com/mandelsoft/kubecrtutils/examples/simple/controllers"
	v1 "k8s.io/api/core/v1"
)

// --- begin definition ---

type Resource = v1.ConfigMap

func Controller() controller.Definition {
	return controller.Define[*Resource, Resource](
		"up",
		controllers.SOURCE,
		support.NewByLogic[*controllers.Options, *controllers.Settings, *Resource, Resource](&ReconcilationLogic{}),
	).
		UseCluster(controllers.TARGET).
		AddTrigger(controller.OwnerTrigger[*Resource]().OnCluster(controllers.TARGET))
}

// --- end definition ---
