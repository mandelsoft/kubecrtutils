package controller

import (
	"context"
	"fmt"

	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/cluster/clustercontext"
	abuilder "github.com/mandelsoft/kubecrtutils/controller/builder"
	mysource "github.com/mandelsoft/kubecrtutils/controller/source"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/mandelsoft/logging"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	sigcluster "sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	multiclusterruntime "sigs.k8s.io/multicluster-runtime"
	mcbuilder "sigs.k8s.io/multicluster-runtime/pkg/builder"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type recorderFunc func(ctx context.Context) record.EventRecorder

type _controller[T any, P kubecrtutils.ObjectPointer[T]] struct {
	controllerManager types.ControllerManager
	definition        TypedDefinition[T, P]
	logger            logging.Logger
	clusters          types.Clusters
	cluster           types.ClusterEquivalent
	gk                schema.GroupKind
	recorder          recorderFunc
	indices           map[string]cacheindex.TypedIndex[T]
	reconciler        reconcile.Reconciler
	watches           mysource.ShadowController[mcreconcile.Request]
}

func (c *_controller[T, P]) GetName() string {
	return c.definition.GetName()
}

func (c *_controller[T, P]) GetFieldManager() string {
	return c.controllerManager.GetName() + "/" + c.definition.GetName()
}

func (c *_controller[T, P]) GetLogger() logging.Logger {
	return c.logger
}

func (c *_controller[T, P]) GetControllerManager() types.ControllerManager {
	return c.controllerManager
}

func (c *_controller[T, P]) GetClusters() types.Clusters {
	return c.clusters
}

func (c *_controller[T, P]) GetResource() client.Object {
	return c.definition.GetResource()
}

func (c *_controller[T, P]) GetDefinition() TypedDefinition[T, P] {
	return c.definition
}

func (c *_controller[T, P]) GetCluster() types.ClusterEquivalent {
	return c.cluster
}

func (c *_controller[T, P]) GetRecoder(ctx context.Context) record.EventRecorder {
	return c.recorder(ctx)
}

func (c *_controller[T, P]) GetTypedIndex(name string) cacheindex.TypedIndex[T] {
	i := c.indices[name]
	return i
}

func (c *_controller[T, P]) GetIndex(name string) cluster.Index {
	return c.GetTypedIndex(name)
}

func (c *_controller[T, P]) GetReconciler() reconcile.Reconciler {
	return c.reconciler
}

func (c *_controller[T, P]) completeCluster(ctx context.Context, cl types.Cluster) error {
	d := c.definition
	mgr := c.GetControllerManager()
	logger := c.GetLogger()

	bldr := ctrl.NewControllerManagedBy(mgr.GetManager().GetLocalManager()).Named(d.GetName())
	mybldr := abuilder.ForCluster(bldr, cl)

	if cl.IsSameAs(mgr.GetMainCluster()) {
		logger.Info("configure reconciling {{kind}} at main cluster {{cluster}}[{{effcluster}}]", "kind", c.gk, "cluster", d.GetCluster(), "effcluster", cl.GetEffective().GetName())
		bldr.For(d.GetResource(), builder.WithPredicates(d.GetWatchPredicates()...))
	} else {
		logger.Info("configure reconciling {{kind}} at cluster {{cluster}}[[[effcluster}}]", "kind", c.gk, "cluster", d.GetCluster(), "effcluster", cl.GetEffective().GetName())
		bldr.WatchesRawSource(
			source.Kind(
				cl.GetCache(),
				d.GetResource(),
				&handler.EnqueueRequestForObject{},
				d.GetWatchPredicates()...,
			),
		)
	}

	trigger, err := cl.TriggerSource(d.GetResource())
	if err != nil {
		return fmt.Errorf("explicit trigger [%s]: %w", c.gk, err)
	}
	logger.Info("configure explicit trigger for main resource {{kind}} at cluster {{cluster}}[{{effcluster}}]", "kind", c.gk, "cluster", d.GetCluster(), "effcluster", c.GetCluster().GetEffective().GetName())
	bldr.WatchesRawSource(trigger)

	logger.Info("configure reconciler")
	r, err := d.GetReconciler().CreateReconciler(ctx, c, mybldr)
	if err != nil {
		return err
	}
	c.reconciler = r

	for _, t := range d.GetTriggers() {
		err := c.addClusterTrigger(bldr, t)
		if err != nil {
			return fmt.Errorf("adding trigger %w", err)
		}
	}

	err = bldr.Complete(&reconcileClusterWrapper[T, P]{c, r})
	if err != nil {
		return err
	}
	return nil
}

func (c *_controller[T, P]) addClusterTrigger(bldr *builder.Builder, tdef ResourceTriggerDefinition) error {
	mgr := c.GetControllerManager()
	gk, err := kubecrtutils.GKForObject(c.GetCluster(), tdef.GetResource())
	if err != nil {
		return fmt.Errorf("cannot determine group kind for %T: %w", tdef.GetResource(), err)
	}

	wc := c.GetClusters().Get(tdef.GetCluster())
	if wc == nil {
		return fmt.Errorf("cluster %q not found", tdef.GetCluster())
	}

	if wc.AsFleet() == nil {
		// single cluster watch
		cl := c.cluster.AsCluster()
		mapper := tdef.GetMapper().SingleTarget(cl, c)

		c.logger.Info("configure resource-based trigger {{resource}}[{{trigger}}] on cluster {{cluster}}[{{effcluster}}]", "trigger", tdef.GetDescription(), "resource", gk, "cluster", tdef.GetCluster(), "effcluster", cl.GetEffective().GetName())
		h := handler.EnqueueRequestsFromMapFunc(mapper)
		if wc.IsSameAs(mgr.GetMainCluster()) {
			bldr.Watches(tdef.GetResource(), h)
		} else {
			bldr.WatchesRawSource(
				source.Kind[client.Object](cl.GetCache(), tdef.GetResource(), h),
			)
		}
	} else {
		// watch is in fleet, but request is on local (non-fleet)
		cl := c.cluster.AsFleet()
		mapper := tdef.GetMapper().MultiTarget(c.cluster.AsCluster(), c)

		if c.watches == nil {
			// create pseudo slave multi cluster controller to host required multi cluster watches.
			c.watches, err = mysource.NewShadowController[mcreconcile.Request](c)
			if err != nil {
				return err
			}
		}
		src, err := mysource.Kind(c.watches, c.cluster, tdef.GetResource(), typedEnqueueRequestsFromMapFunc(mapper))
		if err != nil {
			return err
		}
		c.logger.Info("configure resource-based trigger {{resource}}[{{trigger}}] on fleet {{cluster}}[{{effcluster}}]", "trigger", tdef.GetDescription(), "resource", gk, "cluster", tdef.GetCluster(), "effcluster", cl.GetEffective().GetName())
		bldr = bldr.WatchesRawSource(src)
	}
	return nil
}

func (c *_controller[T, P]) completeFleet(ctx context.Context, cl types.Fleet) error {
	d := c.definition

	mgr := c.GetControllerManager()
	logger := c.GetLogger()

	bldr := multiclusterruntime.NewControllerManagedBy(mgr.GetManager()).Named(d.GetName())

	logger.Info("configure reconciling {{kind}} at fleet {{cluster}}[[[effcluster}}]", "kind", c.gk, "cluster", d.GetCluster(), "effcluster", cl.GetEffective().GetName())
	bldr = bldr.For(d.GetResource(), mcbuilder.WithPredicates(d.GetWatchPredicates()...), mcbuilder.WithClusterFilter(c.GetCluster().Filter))

	trigger, err := cl.TriggerSource(d.GetResource())
	if err != nil {
		return fmt.Errorf("explicit trigger [%s]: %w", c.gk, err)
	}
	logger.Info("configure explicit trigger for main resource {{kind}} at cluster {{cluster}}[{{effcluster}}]", "kind", c.gk, "cluster", d.GetCluster(), "effcluster", c.GetCluster().GetEffective().GetName())
	bldr.WatchesRawSource(trigger)

	logger.Info("configure reconciler")
	r, err := d.GetReconciler().CreateReconciler(ctx, c, abuilder.ForFleet(bldr, c.GetCluster().AsFleet()))
	if err != nil {
		return err
	}
	c.reconciler = r

	for _, t := range d.GetTriggers() {
		err := c.addFleetTrigger(bldr, t)
		if err != nil {
			return fmt.Errorf("resource-based trigger: %w", err)
		}
	}

	err = bldr.Complete(&reconcileFleetWrapper[T, P]{c, r})
	if err != nil {
		return err
	}
	return nil
}

func (c *_controller[T, P]) Complete(ctx context.Context) error {
	cl := c.GetCluster().AsCluster()
	if cl != nil {
		return c.completeCluster(ctx, cl)
	}
	return c.completeFleet(ctx, c.GetCluster().AsFleet())
}

func (c *_controller[T, P]) addFleetTrigger(bldr *mcbuilder.Builder, tdef ResourceTriggerDefinition) error {
	d := c.definition
	mapper := tdef.GetMapper()
	fleet := c.GetCluster().AsFleet()
	gk, err := kubecrtutils.GKForObject(c.GetCluster(), tdef.GetResource())
	if err != nil {
		return fmt.Errorf("cannot determine group kind for %T: %w", tdef.GetResource(), err)
	}

	target := c.GetClusters().Get(tdef.GetCluster())

	if target.AsFleet() == nil {
		// single watch cluster
		cl := target.AsCluster()
		c.logger.Info("configure resource-based trigger for {{resource}}[{{trigger}}] on cluster {{cluster}[{{effcluster}}]}", "trigger", tdef.GetDescription(), "resource", gk, "cluster", d.GetCluster(), "effcluster", cl.GetEffective().GetName())

		src := source.TypedKind[client.Object, mcreconcile.Request](
			target.AsCluster().GetCache(),
			tdef.GetResource(),
			handler.TypedEnqueueRequestsFromMapFunc(mapper.MultiTarget(target.AsCluster(), c)),
		)
		bldr.WatchesRawSource(src)
	} else {
		// multi watch cluster
		creator := func(clusterName string, cl sigcluster.Cluster) handler.TypedEventHandler[client.Object, mcreconcile.Request] {
			return handler.TypedEnqueueRequestsFromMapFunc[client.Object, mcreconcile.Request](mapper.MultiTarget(fleet.GetCluster(clusterName), c))
		}

		cl := target.AsFleet()
		c.logger.Info("configure resource-based trigger for {{resource}}[{{trigger}}] on fleet {{cluster}[{{effcluster}}]}", "trigger", tdef.GetDescription(), "resource", gk, "cluster", d.GetCluster(), "effcluster", cl.GetEffective().GetName())

		bldr.Watches(
			tdef.GetResource(),
			creator,
			mcbuilder.WithClusterFilter(c.GetCluster().Filter),
		)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type reconcileClusterWrapper[T any, P kubecrtutils.ObjectPointer[T]] struct {
	controller Controller[T, P]
	reconciler reconcile.Reconciler
}

type dummy struct {
	client.Object
}

var _ reconcile.TypedReconciler[reconcile.Request] = (*reconcileClusterWrapper[dummy, *dummy])(nil)

func (r *reconcileClusterWrapper[T, P]) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	cl := r.controller.GetCluster().AsCluster()
	return r.reconciler.Reconcile(clustercontext.WithCluster(ctx, cl), request)
}

////////////////////////////////////////////////////////////////////////////////

type reconcileFleetWrapper[T any, P kubecrtutils.ObjectPointer[T]] struct {
	controller Controller[T, P]
	reconciler reconcile.Reconciler
}

var _ reconcile.TypedReconciler[mcreconcile.Request] = (*reconcileFleetWrapper[dummy, *dummy])(nil)

func (r *reconcileFleetWrapper[T, P]) Reconcile(ctx context.Context, request mcreconcile.Request) (reconcile.Result, error) {
	cl := r.controller.GetCluster().AsFleet().GetCluster(request.ClusterName)
	return r.reconciler.Reconcile(clustercontext.WithCluster(ctx, cl), request.Request)
}
