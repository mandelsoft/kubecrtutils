package controller

import (
	"context"
	"fmt"

	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/cluster/clustercontext"
	abuilder "github.com/mandelsoft/kubecrtutils/controller/builder"
	myhandler "github.com/mandelsoft/kubecrtutils/controller/handler"
	"github.com/mandelsoft/kubecrtutils/owner"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/mandelsoft/logging"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	sigcluster "sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	multiclusterruntime "sigs.k8s.io/multicluster-runtime"
	mcbuilder "sigs.k8s.io/multicluster-runtime/pkg/builder"
	mchandler "sigs.k8s.io/multicluster-runtime/pkg/handler"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type recorderFunc func(ctx context.Context) record.EventRecorder

type _controller[P kubecrtutils.ObjectPointer[T], T any] struct {
	controllerManager types.ControllerManager
	definition        TypedDefinition[P, T]
	logger            logging.Logger
	clusters          types.Clusters
	cluster           types.ClusterEquivalent
	gk                schema.GroupKind
	recorder          recorderFunc
	indices           map[string]cacheindex.TypedIndex[T]
	reconciler        reconcile.Reconciler
	ohandler          owner.Handler
}

func (c *_controller[P, T]) GetName() string {
	return c.definition.GetName()
}

func (c *_controller[P, T]) GetOwnerhandler() owner.Handler {
	return c.ohandler
}

func (c *_controller[P, T]) GetFieldManager() string {
	return c.controllerManager.GetName() + "/" + c.definition.GetName()
}

func (c *_controller[P, T]) GetLogger() logging.Logger {
	return c.logger
}

func (c *_controller[P, T]) GetControllerManager() types.ControllerManager {
	return c.controllerManager
}

func (c *_controller[P, T]) GetClusters() types.Clusters {
	return c.clusters
}

func (c *_controller[P, T]) GetResource() client.Object {
	return c.definition.GetResource()
}

func (c *_controller[P, T]) GetDefinition() TypedDefinition[P, T] {
	return c.definition
}

func (c *_controller[P, T]) GetCluster() types.ClusterEquivalent {
	return c.cluster
}

func (c *_controller[P, T]) GetRecoder(ctx context.Context) record.EventRecorder {
	return c.recorder(ctx)
}

func (c *_controller[P, T]) GetTypedIndex(name string) cacheindex.TypedIndex[T] {
	i := c.indices[name]
	return i
}

func (c *_controller[P, T]) GetIndex(name string) cluster.Index {
	return c.GetTypedIndex(name)
}

func (c *_controller[P, T]) GetReconciler() reconcile.Reconciler {
	return c.reconciler
}

func (c *_controller[P, T]) Complete(ctx context.Context) error {
	cl := c.GetCluster()
	d := c.definition

	mgr := c.GetControllerManager()
	logger := c.GetLogger()

	bldr := multiclusterruntime.NewControllerManagedBy(mgr.GetManager()).Named(d.GetName())

	logger.Info("configure reconciling {{kind}} at {{type}} {{cluster}}[[[effcluster}}]", "kind", c.gk, "type", cl.GetTypeInfo(), "cluster", d.GetCluster(), "effcluster", cl.GetEffective().GetName())
	bldr = bldr.For(d.GetResource(), mcbuilder.WithPredicates(d.GetWatchPredicates()...), mcbuilder.WithClusterFilter(c.GetCluster().Filter))

	trigger, err := cl.TriggerSource(d.GetResource())
	if err != nil {
		return fmt.Errorf("explicit trigger [%s]: %w", c.gk, err)
	}
	logger.Info("configure explicit trigger for main resource {{kind}} at cluster {{cluster}}[{{effcluster}}]", "kind", c.gk, "cluster", d.GetCluster(), "effcluster", c.GetCluster().GetEffective().GetName())
	bldr.WatchesRawSource(trigger)

	logger.Info("configure reconciler")
	r, err := d.GetReconciler().CreateReconciler(ctx, c, abuilder.For(bldr, c.GetCluster()))
	if err != nil {
		return err
	}
	c.reconciler = r

	for _, t := range d.GetTriggers() {
		err := c.addTrigger(ctx, bldr, t)
		if err != nil {
			return fmt.Errorf("resource-based trigger: %w", err)
		}
	}

	err = bldr.Complete(&reconcileWrapper[P, T]{c, r})
	if err != nil {
		return err
	}
	return nil
}

func (c *_controller[P, T]) addTrigger(ctx context.Context, bldr *mcbuilder.Builder, tdef ResourceTriggerDefinition) error {
	d := c.definition
	gk, err := kubecrtutils.GKForObject(c.GetCluster(), tdef.GetResource())
	if err != nil {
		return fmt.Errorf("cannot determine group kind for %T: %w", tdef.GetResource(), err)
	}

	target := c.GetClusters().Get(tdef.GetCluster())

	c.logger.Info("configure resource-based trigger for {{resource}}[{{trigger}}] on {{type}} {{cluster}}[{{effcluster}}]}", "trigger", tdef.GetDescription(), "resource", gk, "type", target.GetTypeInfo(), "cluster", d.GetCluster(), "effcluster", target.GetEffective().GetName())

	bldr.Watches(
		tdef.GetResource(),
		watchWrapper[P, T](c, myhandler.EnqueueRequestFromMapFuncFactory(tdef.GetMapper()(ctx, c))),
		mcbuilder.WithClusterFilter(target.Filter),
	)
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type reconcileWrapper[P kubecrtutils.ObjectPointer[T], T any] struct {
	controller Controller[P, T]
	reconciler reconcile.Reconciler
}

func (r *reconcileWrapper[P, T]) Reconcile(ctx context.Context, request mcreconcile.Request) (reconcile.Result, error) {
	cl := r.controller.GetControllerManager().MapTechnicalName(request.ClusterName).AsCluster()
	return r.reconciler.Reconcile(clustercontext.WithCluster(ctx, cl), request.Request)
}

////////////////////////////////////////////////////////////////////////////////

func watchWrapper[P kubecrtutils.ObjectPointer[T], T any](controller Controller[P, T], factory mchandler.EventHandlerFunc) mchandler.EventHandlerFunc {
	return func(clusterName string, cluster sigcluster.Cluster) mchandler.EventHandler {
		cl := controller.GetControllerManager().MapTechnicalName(clusterName).AsCluster()
		return &wrapperHandler{cl, factory(clusterName, cluster)}
	}
}

type wrapperHandler struct {
	cluster types.Cluster
	handler handler.TypedEventHandler[client.Object, mcreconcile.Request]
}

var _ handler.TypedEventHandler[client.Object, mcreconcile.Request] = (*wrapperHandler)(nil)

func (w *wrapperHandler) Create(ctx context.Context, e event.TypedCreateEvent[client.Object], wq workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	w.handler.Create(clustercontext.WithCluster(ctx, w.cluster), e, wq)
}

func (w *wrapperHandler) Update(ctx context.Context, e event.TypedUpdateEvent[client.Object], wq workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	w.handler.Update(clustercontext.WithCluster(ctx, w.cluster), e, wq)
}

func (w *wrapperHandler) Delete(ctx context.Context, e event.TypedDeleteEvent[client.Object], wq workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	w.handler.Delete(clustercontext.WithCluster(ctx, w.cluster), e, wq)
}

func (w *wrapperHandler) Generic(ctx context.Context, e event.TypedGenericEvent[client.Object], wq workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	w.handler.Generic(clustercontext.WithCluster(ctx, w.cluster), e, wq)
}
