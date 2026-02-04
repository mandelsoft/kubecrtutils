package helper

import (
	"time"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

func MapQueueMCtoCR(clusterName string, q workqueue.TypedRateLimitingInterface[mcreconcile.Request]) workqueue.TypedRateLimitingInterface[reconcile.Request] {
	return &queueMCtoCR{
		TypedRateLimitingInterface: q,
		cluster:                    clusterName,
	}
}

type queueMCtoCR struct {
	workqueue.TypedRateLimitingInterface[mcreconcile.Request]
	cluster string
}

var _ workqueue.TypedRateLimitingInterface[reconcile.Request] = (*queueMCtoCR)(nil)

func (q *queueMCtoCR) Add(item reconcile.Request) {
	q.TypedRateLimitingInterface.Add(mcreconcile.Request{ClusterName: q.cluster, Request: item})
}

func (q *queueMCtoCR) Get() (item reconcile.Request, shutdown bool) {
	i, s := q.TypedRateLimitingInterface.Get()
	return i.Request, s
}

func (q *queueMCtoCR) Done(item reconcile.Request) {
	q.TypedRateLimitingInterface.Done(mcreconcile.Request{ClusterName: q.cluster, Request: item})
}

func (q *queueMCtoCR) AddAfter(item reconcile.Request, duration time.Duration) {
	q.TypedRateLimitingInterface.AddAfter(mcreconcile.Request{ClusterName: q.cluster, Request: item}, duration)
}

func (q *queueMCtoCR) AddRateLimited(item reconcile.Request) {
	q.TypedRateLimitingInterface.AddRateLimited(mcreconcile.Request{ClusterName: q.cluster, Request: item})
}

func (q *queueMCtoCR) Forget(item reconcile.Request) {
	q.TypedRateLimitingInterface.Forget(mcreconcile.Request{ClusterName: q.cluster, Request: item})
}

func (q *queueMCtoCR) NumRequeues(item reconcile.Request) int {
	return q.TypedRateLimitingInterface.NumRequeues(mcreconcile.Request{ClusterName: q.cluster, Request: item})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func MapQueueCRtoMC(clusterName string, q workqueue.TypedRateLimitingInterface[reconcile.Request]) workqueue.TypedRateLimitingInterface[mcreconcile.Request] {
	return &queueCRtoMC{
		TypedRateLimitingInterface: q,
		cluster:                    clusterName,
	}
}

type queueCRtoMC struct {
	workqueue.TypedRateLimitingInterface[reconcile.Request]
	cluster string
}

var _ workqueue.TypedRateLimitingInterface[mcreconcile.Request] = (*queueCRtoMC)(nil)

func (q *queueCRtoMC) filter(item mcreconcile.Request) bool {
	return item.ClusterName == q.cluster || item.ClusterName == ""
}

func (q *queueCRtoMC) Add(item mcreconcile.Request) {
	if q.filter(item) {
		q.TypedRateLimitingInterface.Add(item.Request)
	}
}

func (q *queueCRtoMC) Get() (item mcreconcile.Request, shutdown bool) {
	i, s := q.TypedRateLimitingInterface.Get()
	return mcreconcile.Request{Request: i, ClusterName: q.cluster}, s
}

func (q *queueCRtoMC) Done(item mcreconcile.Request) {
	if q.filter(item) {
		q.TypedRateLimitingInterface.Done(item.Request)
	}
}

func (q *queueCRtoMC) AddAfter(item mcreconcile.Request, duration time.Duration) {
	if q.filter(item) {
		q.TypedRateLimitingInterface.AddAfter(item.Request, duration)
	}
}

func (q *queueCRtoMC) AddRateLimited(item mcreconcile.Request) {
	if q.filter(item) {
		q.TypedRateLimitingInterface.AddRateLimited(item.Request)
	}
}

func (q *queueCRtoMC) Forget(item mcreconcile.Request) {
	if q.filter(item) {
		q.TypedRateLimitingInterface.Forget(item.Request)
	}
}

func (q *queueCRtoMC) NumRequeues(item mcreconcile.Request) int {
	if q.filter(item) {
		return q.TypedRateLimitingInterface.NumRequeues(item.Request)
	}
	return 0
}
