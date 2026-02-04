package controller

import (
	"context"
	"slices"
	"sync"

	"github.com/mandelsoft/kubecrtutils/controller/helper"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	mchandler "sigs.k8s.io/multicluster-runtime/pkg/handler"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
	mcsource "sigs.k8s.io/multicluster-runtime/pkg/source"
)

var _ mcsource.Source

type kind struct {
	mu sync.Mutex

	manager ShadowController[mcreconcile.Request]

	cluster    types.ClusterEquivalent
	object     client.Object
	predicates []predicate.TypedPredicate[client.Object]
	handler    mchandler.EventHandlerFunc

	factory mcsource.TypedSource[client.Object, mcreconcile.Request]
	source  source.SyncingSource
	queue   workqueue.TypedRateLimitingInterface[reconcile.Request]
}

// usable for simple cluster controllers
var _ source.TypedSource[reconcile.Request] = (*kind)(nil)

// usable for fleet controllers
var _ mcsource.TypedSource[client.Object, mcreconcile.Request] = (*kind)(nil)

// Kind is used for multi cluster watches for crt controllers.
// It requires a pseudo multi cluster controller hosting the muti cluster watches.
// It acts as raw source for then crt controller and multi-cluster watch on
// the pseudo controller.
func Kind(
	manager ShadowController[mcreconcile.Request],
	cluster types.ClusterEquivalent,
	obj client.Object,
	handler mchandler.EventHandlerFunc,
	predicates ...predicate.TypedPredicate[client.Object]) (*kind, error) {
	k := &kind{
		cluster:    cluster,
		handler:    handler,
		object:     obj,
		predicates: slices.Clone(predicates),
	}
	if cluster.AsFleet() != nil {
		err := manager.MultiClusterWatch(k)
		if err != nil {
			return nil, err
		}
	}
	return k, nil
}

// Start is used for regular clusters managed by the local manager
func (k *kind) Start(ctx context.Context, w workqueue.TypedRateLimitingInterface[reconcile.Request]) error {
	k.queue = w
	// is source is fleet, watch start is delayed to cluster engagement calling ForCluster
	if k.cluster.AsFleet() != nil {
		return nil
	}
	// source is no fleet, so establish regular watch

	h := k.handler(k.cluster.GetName(), k.cluster.AsCluster().GetCluster())
	k.source = source.Kind[client.Object](k.cluster.AsCluster().GetCache(), k.object, helper.MapHandlerMCtoCR(k.cluster.AsCluster(), h), k.predicates...)
	return k.source.Start(ctx, w)
}

// ForCluster is called when source cluster is a fleet.
func (k *kind) ForCluster(clusterName string, cluster cluster.Cluster) (source.TypedSource[mcreconcile.Request], bool, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.factory == nil {
		if k.cluster.AsFleet() == nil {
			k.factory = mcsource.TypedKind[client.Object, mcreconcile.Request](k.object, redirectingEventHandlerFunc(k), k.predicates...).WithClusterFilter(k.cluster.Filter)
		} else {
			k.factory = mcsource.TypedKind[client.Object, mcreconcile.Request](k.object, k.handler, k.predicates...).WithClusterFilter(k.cluster.Filter)
		}
	}
	return k.factory.ForCluster(clusterName, cluster)
}

func redirectingEventHandlerFunc(k *kind) mchandler.TypedEventHandlerFunc[client.Object, mcreconcile.Request] {
	return func(clusterName string, cluster cluster.Cluster) mchandler.EventHandler {
		h := k.handler(clusterName, cluster)
		return &redirectingHandler{clusterName: clusterName, kind: k, handler: h}
	}
}

type redirectingHandler struct {
	clusterName string
	kind        *kind
	handler     mchandler.EventHandler
}

var _ handler.TypedEventHandler[client.Object, mcreconcile.Request] = (*redirectingHandler)(nil)

func (r *redirectingHandler) Create(ctx context.Context, e event.TypedCreateEvent[client.Object], _ workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	r.handler.Create(ctx, e, helper.MapQueueCRtoMC(r.clusterName, r.kind.queue))
}

func (r *redirectingHandler) Update(ctx context.Context, e event.TypedUpdateEvent[client.Object], _ workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	r.handler.Update(ctx, e, helper.MapQueueCRtoMC(r.clusterName, r.kind.queue))
}

func (r *redirectingHandler) Delete(ctx context.Context, e event.TypedDeleteEvent[client.Object], _ workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	r.handler.Delete(ctx, e, helper.MapQueueCRtoMC(r.clusterName, r.kind.queue))
}

func (r *redirectingHandler) Generic(ctx context.Context, e event.TypedGenericEvent[client.Object], _ workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	r.handler.Generic(ctx, e, helper.MapQueueCRtoMC(r.clusterName, r.kind.queue))
}
