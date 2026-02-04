package helper

import (
	"context"

	"github.com/mandelsoft/kubecrtutils/cluster/clustercontext"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mchandler "sigs.k8s.io/multicluster-runtime/pkg/handler"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type handlerCRtoMC struct {
	handler handler.TypedEventHandler[client.Object, reconcile.Request]
	cluster string
}

func MapHandlerCRtoMC(clusterName string, eventHandler handler.TypedEventHandler[client.Object, reconcile.Request]) mchandler.EventHandler {

	return &handlerCRtoMC{
		handler: eventHandler,
		cluster: clusterName,
	}
}

var _ handler.TypedEventHandler[client.Object, mcreconcile.Request] = (*handlerCRtoMC)(nil)

func (h *handlerCRtoMC) Create(ctx context.Context, e event.TypedCreateEvent[client.Object], w workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	h.handler.Create(ctx, e, MapQueueMCtoCR(h.cluster, w))
}

func (h *handlerCRtoMC) Update(ctx context.Context, e event.TypedUpdateEvent[client.Object], w workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	h.handler.Update(ctx, e, MapQueueMCtoCR(h.cluster, w))
}

func (h *handlerCRtoMC) Delete(ctx context.Context, e event.TypedDeleteEvent[client.Object], w workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	h.handler.Delete(ctx, e, MapQueueMCtoCR(h.cluster, w))
}

func (h *handlerCRtoMC) Generic(ctx context.Context, e event.TypedGenericEvent[client.Object], w workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	h.handler.Generic(ctx, e, MapQueueMCtoCR(h.cluster, w))
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type handlerMCtoCR struct {
	handler handler.TypedEventHandler[client.Object, mcreconcile.Request]
	cluster types.Cluster
}

func MapHandlerMCtoCR(cluster types.Cluster, eventHandler handler.TypedEventHandler[client.Object, mcreconcile.Request]) handler.EventHandler {

	return &handlerMCtoCR{
		handler: eventHandler,
		cluster: cluster,
	}
}

var _ handler.TypedEventHandler[client.Object, mcreconcile.Request] = (*handlerCRtoMC)(nil)

func (h *handlerMCtoCR) Create(ctx context.Context, e event.TypedCreateEvent[client.Object], w workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	clustercontext.WithCluster(ctx, h.cluster)
	h.handler.Create(ctx, e, MapQueueCRtoMC(h.cluster.GetName(), w))
}

func (h *handlerMCtoCR) Update(ctx context.Context, e event.TypedUpdateEvent[client.Object], w workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	h.handler.Update(ctx, e, MapQueueCRtoMC(h.cluster.GetName(), w))
}

func (h *handlerMCtoCR) Delete(ctx context.Context, e event.TypedDeleteEvent[client.Object], w workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	h.handler.Delete(ctx, e, MapQueueCRtoMC(h.cluster.GetName(), w))
}

func (h *handlerMCtoCR) Generic(ctx context.Context, e event.TypedGenericEvent[client.Object], w workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	h.handler.Generic(ctx, e, MapQueueCRtoMC(h.cluster.GetName(), w))
}

////////////////////////////////////////////////////////////////////////////////

func AnnotateMapFuncWithCluster(clusters types.Clusters, handlerFunc mchandler.EventHandlerFunc) mchandler.EventHandlerFunc {
	return func(clusterName string, cluster cluster.Cluster) mchandler.EventHandler {
		h := handlerFunc(clusterName, cluster)
		return &annotatedHandler{clusters.Get(clusterName).AsCluster(), h}
	}
}

func AnnotateHandler(cluster types.Cluster, handler mchandler.EventHandler) mchandler.EventHandler {
	return &annotatedHandler{cluster, handler}
}

type annotatedHandler struct {
	cluster types.Cluster
	handler mchandler.EventHandler
}

var _ mchandler.EventHandler = (*annotatedHandler)(nil)

func (a *annotatedHandler) Create(ctx context.Context, e event.TypedCreateEvent[client.Object], w workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	a.handler.Create(clustercontext.WithCluster(ctx, a.cluster), e, w)
}

func (a *annotatedHandler) Update(ctx context.Context, e event.TypedUpdateEvent[client.Object], w workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	a.handler.Update(clustercontext.WithCluster(ctx, a.cluster), e, w)
}

func (a *annotatedHandler) Delete(ctx context.Context, e event.TypedDeleteEvent[client.Object], w workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	a.handler.Delete(clustercontext.WithCluster(ctx, a.cluster), e, w)
}

func (a *annotatedHandler) Generic(ctx context.Context, e event.TypedGenericEvent[client.Object], w workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	a.handler.Generic(clustercontext.WithCluster(ctx, a.cluster), e, w)
}
