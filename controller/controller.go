package controller

import (
	"context"
	"fmt"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/cluster/clustercontext"
	abuilder "github.com/mandelsoft/kubecrtutils/controller/builder"
	. "github.com/mandelsoft/kubecrtutils/log"
	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/objutils"
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

type FinalizerModifier interface {
	ModifyFinalizer(s string) string
}

type _controller[P kubecrtutils.ObjectPointer[T], T any] struct {
	enforceNameExtension bool
	controllerManager    types.ControllerManager
	definition           TypedDefinition[P, T]
	logger               logging.Logger
	mappings             mapping.Mappings // cluster mappings
	components           types.Components
	clusters             types.Clusters
	cluster              types.ClusterEquivalent
	gk                   schema.GroupKind
	recorder             recorderFunc
	localIndices         map[string]cacheindex.TypedIndex[T]
	allIndices           map[string]cacheindex.Index
	reconciler           reconcile.Reconciler
	ohandler             owner.Handler
	finalizer            string
}

func (c *_controller[P, T]) GetName() string {
	return c.definition.GetName()
}

func (c *_controller[P, T]) GetOptions() flagutils.Options {
	return c.definition.GetOptions()
}

func (c *_controller[P, T]) GetOwnerHandler() owner.Handler {
	return c.ohandler
}

func (c *_controller[P, T]) GetFieldManager() string {
	return c.controllerManager.GetName() + "/" + c.definition.GetName()
}

func (c *_controller[P, T]) GetFinalizer() string {
	return c.finalizer
}

func (c *_controller[P, T]) GetLogger() logging.Logger {
	return c.logger
}

func (c *_controller[P, T]) GetControllerManager() types.ControllerManager {
	return c.controllerManager
}

func (c *_controller[P, T]) GetClusterMappings() mapping.Mappings {
	return c.mappings
}

func (c *_controller[P, T]) GetClusters() types.Clusters {
	return c.clusters
}

func (c *_controller[P, T]) GetComponents() types.Components {
	return c.components
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

func (c *_controller[P, T]) GetLogicalCluster(name string) types.ClusterEquivalent {
	return c.clusters.Get(name)
}

func (c *_controller[P, T]) GetRecoder(ctx context.Context) record.EventRecorder {
	return c.recorder(ctx)
}

func (c *_controller[P, T]) GetLocalIndex(name string) cacheindex.TypedIndex[T] {
	i := c.localIndices[name]
	return i
}

func (c *_controller[P, T]) GetIndex(name string) cluster.Index {
	return c.allIndices[name]
}

func (c *_controller[P, T]) GetReconciler() reconcile.Reconciler {
	return c.reconciler
}

func (c *_controller[P, T]) Complete(ctx context.Context) error {
	cl := c.GetCluster()
	d := c.definition

	mgr := c.GetControllerManager()
	logger := c.GetLogger()

	Info(logger, "- complete controller {{controller}}", "controller", c.GetName())

	bldr := multiclusterruntime.NewControllerManagedBy(mgr.GetManager()).Named(d.GetName())

	Info(logger, "  configure reconciling of ", GroupKind(c.gk), " at ", LogicalClusterInfo(c.GetCluster()))
	bldr = bldr.For(d.GetResource(), mcbuilder.WithPredicates(d.GetWatchPredicates()...), mcbuilder.WithClusterFilter(c.GetCluster().Filter))

	trigger, err := cl.TriggerSource(d.GetResource())
	if err != nil {
		return fmt.Errorf("explicit trigger [%s]: %w", c.gk, err)
	}
	Info(logger, "  configure explicit trigger for main resource ", GroupKind(c.gk), " at ", LogicalClusterInfo(c.GetCluster()))
	bldr.WatchesRawSource(trigger)

	logger.Info("  configure reconciler")
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

	err = bldr.Complete(&reconcileWrapper[P, T]{cl, r})
	if err != nil {
		return err
	}
	return nil
}

func (c *_controller[P, T]) addTrigger(ctx context.Context, bldr *mcbuilder.Builder, tdef ResourceTriggerDefinition) error {
	gk, err := kubecrtutils.GKForObject(c.GetCluster(), tdef.GetResource())
	if err != nil {
		return fmt.Errorf("cannot determine group kind for %T: %w", tdef.GetResource(), err)
	}

	cname := general.OptionalNonZeroDefaulted(c.definition.GetCluster(), tdef.GetCluster())
	target := c.GetClusters().Get(cname)
	if target == nil {
		return fmt.Errorf("cannot determine target cluster for %q: %w", cname)
	}
	Info(c.logger, "  configure resource-based trigger for ", GroupKind(gk), "[", KeyValue("description", tdef.GetDescription()), "] on ", LogicalClusterInfo(target))

	m, err := tdef.GetMapper()(ctx, c)
	if err != nil {
		return fmt.Errorf("trigger %q: %w", err)
	}
	bldr.Watches(
		tdef.GetResource(),
		watchWrapper[P, T](target, enqueueRequestFromMapFuncFactory(m), tdef),
		mcbuilder.WithClusterFilter(target.Filter),
	)
	return nil
}

func (c *_controller[P, T]) GenerateNameFor(ctx context.Context, tgt types.Cluster, prefix, namespace, name string, len ...int) string {
	src := clustercontext.ClusterFor(ctx)
	if src == nil || tgt.IsSameAs(src) {
		return objutils.GenerateUniqueName(prefix, "", namespace, name, len...)
	}
	if c.enforceNameExtension || c.cluster.AsFleet() != nil || c.controllerManager.GetClusters().Len() > 2 {
		return objutils.GenerateUniqueName(prefix, src.GetName(), namespace, name, len...)
	}
	return objutils.GenerateUniqueName(prefix, "", namespace, name, len...)
}

////////////////////////////////////////////////////////////////////////////////

type reconcileWrapper[P kubecrtutils.ObjectPointer[T], T any] struct {
	cluster    types.ClusterEquivalent
	reconciler reconcile.Reconciler
}

func (r *reconcileWrapper[P, T]) Reconcile(ctx context.Context, request mcreconcile.Request) (reconcile.Result, error) {
	// we propagate the cluster with its logical name.as defined by the controller definition
	n, cl := r.cluster.LiftTechnical(request.ClusterName)
	// handle vanished cluster engagement by propagating name separately
	return r.reconciler.Reconcile(clustercontext.WithClusterAndName(ctx, cl, n), request.Request)
}

////////////////////////////////////////////////////////////////////////////////

func watchWrapper[P kubecrtutils.ObjectPointer[T], T any](target types.ClusterEquivalent, factory mchandler.EventHandlerFunc, def ResourceTriggerDefinition) mchandler.EventHandlerFunc {
	return func(clusterName string, cluster sigcluster.Cluster) mchandler.EventHandler {
		n, cl := target.LiftTechnical(clusterName)
		if cl == nil {
			// return handler to avoid crashed for omitted clusters
		}
		return &wrapperHandler{cl, n, factory(clusterName, cluster)}
	}
}

type wrapperHandler struct {
	cluster     types.Cluster
	clusterName string
	handler     handler.TypedEventHandler[client.Object, mcreconcile.Request]
}

var _ handler.TypedEventHandler[client.Object, mcreconcile.Request] = (*wrapperHandler)(nil)

func (w *wrapperHandler) setContext(ctx context.Context) context.Context {
	return clustercontext.WithClusterAndName(ctx, w.cluster, w.clusterName)
}

func (w *wrapperHandler) Create(ctx context.Context, e event.TypedCreateEvent[client.Object], wq workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	w.handler.Create(w.setContext(ctx), e, wq)
}

func (w *wrapperHandler) Update(ctx context.Context, e event.TypedUpdateEvent[client.Object], wq workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	w.handler.Update(w.setContext(ctx), e, wq)
}

func (w *wrapperHandler) Delete(ctx context.Context, e event.TypedDeleteEvent[client.Object], wq workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	w.handler.Delete(w.setContext(ctx), e, wq)
}

func (w *wrapperHandler) Generic(ctx context.Context, e event.TypedGenericEvent[client.Object], wq workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	w.handler.Generic(w.setContext(ctx), e, wq)
}
