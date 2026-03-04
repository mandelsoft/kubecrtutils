package hostedzone

import (
	"context"
	"time"

	"github.com/mandelsoft/kubecrtutils/controller"
	"github.com/mandelsoft/kubecrtutils/controller/builder"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconciler"
	corednsv1alpha1 "github.com/mandelsoft/kubedns/api/coredns/v1alpha1"
	crtreconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReconcilerFactory struct {
	Options
}

func (f *ReconcilerFactory) CreateReconciler(ctx context.Context, controller controller.TypedController[*corednsv1alpha1.HostedZone, corednsv1alpha1.HostedZone], b builder.Builder) (crtreconcile.Reconciler, error) {
	logger := controller.GetLogger()
	logger.Info("  creating hostedzone reconciler...")

	r := &Reconciler{
		controller: controller,
		Runtime:    controller.GetLogicalCluster("runtime"),
		Dataplane:  controller.GetLogicalCluster("dataplane"),
	}

	return reconciler.CRTReconcilerFor(controller, r, 300*time.Second), nil
}
